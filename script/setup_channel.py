# encoding: utf-8

import sys
import subprocess
import yaml

import up


config_file = 'fabric-config.yml'
with open(config_file, 'r') as f:
    config = yaml.load(f)


def check_peer(peer_name):
    cmd = 'docker ps'
    result = subprocess.check_output(cmd, shell=True)
    lines = result.splitlines()[1:]
    containers = [line.split()[-1] for line in lines]

    return peer_name in containers


def make_channel(peer_name):
    peer = config['fabric']['peers'][peer_name]
    e_options_str = ' '.join(['-e "' + e + '"' for e in peer['environment']])

    print '### {}にチャンネルを作成'.format(peer_name)
    cmd_params = {
        'option': e_options_str,
        'container': peer_name,
        'orderer': config['fabric']['orderer']['host'] + ':' + str(config['fabric']['orderer']['port']),
        'channel': config['fabric']['channel'],
        'file': peer['configtx_file'],
    }
    cmd = 'docker exec {option} {container} peer channel create -o {orderer} -c {channel} -f {file}'.format(**cmd_params)
    print cmd
    subprocess.check_call(cmd, shell=True)

    print '### チャンネルに参加'
    cmd_params = {
        'option': e_options_str,
        'container': peer_name,
        'block': config['fabric']['channel'] + '.block',
    }
    cmd = 'docker exec {option} {container} peer channel join -b {block}'.format(**cmd_params)
    print cmd
    subprocess.check_call(cmd, shell=True)


if __name__ == '__main__':
    args = sys.argv

    # help指定ならusageを表示する
    if len(args) > 1 and (args[1] == '--help' or args[1] == '-h'):
        print 'usage: [対話] python setup_channel.py'
        print '     : [指定] python setup_channel.py target_peer'
        exit()

    peers = config['fabric']['peers'].keys()

    if len(args) == 1:
        # 指定がなければ選択式で起動する
        while True:
            print 'チャンネルを設定したいpeerの番号を選択してください。'
            for i, p in enumerate(peers):
                print '[{}] {}'.format(i + 1, p)
            num = int(raw_input('> ')) - 1
            if 0 <= num < len(peers):
                break
            print '[Error] 番号が正しくありません。'
        target = peers[num]
    else:
        # 指定があればそれを利用する
        target = args[1]
        if target not in peers:
            print '[Error] peer「{}」はconfigに記述がありません！'.format(target)
            exit()

    # peerの起動確認と起動
    if not check_peer(target):
        print 'peer「{}」が起動していません。'.format(target)
        yes = raw_input('起動しますか？ [y] > ')
        if yes != 'y':
            exit()
        print '### コンテナ{}を起動'.format(target)
        up.up_container(target)

    # チャンネルを作成する
    # TODO すでにチャンネルが作成済みでjoinするのみである場合を考慮
    make_channel(target)
