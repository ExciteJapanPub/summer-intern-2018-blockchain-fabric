# encoding: utf-8

import os.path
import shutil
import subprocess
import yaml


print '### 確認'
yes = raw_input('実行するとすべてのコンテナ/ネットワークが削除されます。よろしいですか？ [y] > ')
if yes != 'y':
    exit()

config_file = 'fabric-config.yml'
with open(config_file, 'r') as f:
    config = yaml.load(f)

print '### keystore削除'
keystore_path = config['fabric']['keystore_path']
if os.path.exists(keystore_path):
    shutil.rmtree(keystore_path)

print '### 起動中のコンテナを停止'
cmd = 'docker-compose -f ' + config['docker']['compose'] + ' down'
print cmd
subprocess.check_call(cmd, shell=True)

print '### コンテナを削除'
cmd = 'docker ps -aq | wc -l'
result = subprocess.check_output(cmd, shell=True)
if result.strip() != "0":
    cmd = 'docker rm -f $(docker ps -aq)'
    print cmd
    subprocess.check_call(cmd, shell=True)

print '### 使用していないネットワークを削除'
cmd = 'echo y | docker network prune'
print cmd
subprocess.check_call(cmd, shell=True)
print

print '### チェーンコードのdockerイメージを削除'
cmd = 'docker images "dev-*" -q | wc -l'
result = subprocess.check_output(cmd, shell=True)
if result.strip() != "0":
    cmd = 'docker rmi $(docker images "dev-*" -q)'
    print cmd
    subprocess.check_call(cmd, shell=True)
