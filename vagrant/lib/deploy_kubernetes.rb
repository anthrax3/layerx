require_relative './helpers'
require 'yaml'
require 'fileutils'

def get_kubeconfig conf
  source_cfg = "#{ENV['HOME']}/.kube/config"
  destdir = "#{Dir.pwd}/minikube/"
  FileUtils.mkdir_p destdir

  kubecfg = YAML.load_file(source_cfg)
  ca_file = nil
  kubecfg['clusters'].each do |cluster|
    if cluster['name'] == 'minkube'
      ca_file = cluster['cluster']['certificate-authority']
      cluster['cluster']['certificate-authority'] = "#{conf['kube_cfg_home']}/ca.crt"
      break
    end
  end
  client_certificate = nil
  client_key = nil
  kubecfg['users'].each do |user|
    if user['name'] == 'minikube'
      client_certificate = user['user']['client-certificate']
      client_key = user['user']['client-key']
      user['user']['client-certificate'] = "#{conf['kube_cfg_home']}/apiserver.crt"
      user['user']['client-key'] = "#{conf['kube_cfg_home']}/apiserver.key"
    end
  end

  FileUtils.cp(ca_file, "#{destdir}/ca.crt")
  FileUtils.cp(client_certificate, "#{destdir}/apiserver.crt")
  FileUtils.cp(client_key, "#{destdir}/apiserver.key")
  File.open("#{destdir}/kubeconfig", 'w') {|f| f.write kubecfg.to_yaml }
end

def deploy_kubernetes conf
  puts "deploying kubernetes"
  system("minikube start --host-only-cidr=\"#{conf['k8s_ip_base']}1/24\"")
  get_kubeconfig conf
end

def stop_kubernetes
  puts "stopping kubernetes"
  system("minikube stop")
end

def destroy_kubernetes
  puts "destroy kubernetes"
  system("minikube delete")
end

def reload_kubernetes conf
  puts "reloading kubernetes"
  destroy_kubernetes
  deploy_kubernetes conf
end

def status_kubernetes
  system("minikube status")
end

def handle_kubernetes_action conf
  unless in_path?("minikube")
    puts "please install minikube before deploying kubernetes (make sure it is in PATH)"
    exit(1)
  end

  case ARGV.first
    when "up"
      deploy_kubernetes conf
    when "down"
      stop_kubernetes
    when "reload"
      reload_kubernetes conf
    when "destroy"
      destroy_kubernetes
    when "status"
      status_kubernetes
  end
end
