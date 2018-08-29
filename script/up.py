# encoding: utf-8

import sys
import subprocess
import yaml


config_file = 'fabric-config.yml'
with open(config_file, 'r') as f:
    config = yaml.load(f)


def up_all():
    cmd = 'docker-compose -f ' + config['docker']['compose'] + ' up -d'
    print cmd
    subprocess.check_call(cmd, shell=True)


def up_container(container):
    cmd = 'docker-compose -f ' + config['docker']['compose'] + ' up -d ' + container
    print cmd
    subprocess.check_call(cmd, shell=True)


def get_containers():
    with open(config['docker']['compose'], 'r') as f:
        docker_compose = yaml.load(f)
    return docker_compose['services'].keys()


if __name__ == '__main__':
    args = sys.argv

    # help指定ならusageを表示する
    if len(args) > 1 and (args[1] == '--help' or args[1] == '-h'):
        print 'usage: [対話] python up.py'
        print '     : [全て] python up.py --all'
        print '     : [指定] python up.py target1 target2 ...'
        exit()

    # all指定なら全部起動する
    if len(args) > 1 and args[1] == '--all':
        print '### すべてのコンテナを起動'
        up_all()
        exit()

    containers = get_containers()

    # 指定がなければ選択式で起動する
    if len(args) == 1:
        print '起動したいコンテナの番号を選択してください。(複数選択可能スペース区切り)'
        for i, c in enumerate(containers):
            print '[{}] {}'.format(i + 1, c)
        nums = raw_input('> ')
        nums = map(int, nums.split())
        targets = [c for i, c in enumerate(containers) if i + 1 in nums]
        for target in targets:
            print '### コンテナ{}を起動'.format(target)
            up_container(target)
        exit()

    # 指定されているコンテナを起動する
    targets = args[1:]
    for target in targets:
        print '### コンテナ{}を起動'.format(target)
        if target not in containers:
            print '[Error] コンテナ「{}」は{}に記述がありません！'.format(target, config['docker']['compose'])
            continue
        up_container(target)
