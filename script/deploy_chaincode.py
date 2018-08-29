# encoding: utf-8

from datetime import datetime
import sys
import subprocess
import yaml

import up


config_file = 'fabric-config.yml'
with open(config_file, 'r') as f:
    config = yaml.load(f)

clients = config['fabric']['clients'].keys()


def show_usage():
    print 'usage: [対話] python deploy_chaincode.py'
    print '     : [指定] python deploy_chaincode.py target_client chaincode'


def check_client(client_name):
    cmd = 'docker ps'
    result = subprocess.check_output(cmd, shell=True)
    lines = result.splitlines()[1:]
    containers = [line.split()[-1] for line in lines]

    return client_name in containers


def deploy_chaincode(client_name, chaincode_name):
    client = config['fabric']['clients'][client_name]
    e_options_str = ' '.join(['-e "' + e + '"' for e in client['environment']])
    version = datetime.now().strftime('v%Y%m%d%H%M%S')
    # ひとまずGo固定。nodeも選べるようにしたければ拡張
    chaincode_path = 'github.com/{}/go'.format(chaincode_name)
    # ひとまずinit呼び出し時の引数は空
    # initで空以外の引数を受け取るchanincodeを実装するのであれば、実行時オプションで指定できるようにする
    chaincode_initialize_args = '{"Args":[]}'
    # ひとまずポリシーはOrg1MSPのみでの承認。変えたい場合は実行時オプションで指定できるようにする
    policy = 'OR ("Org1MSP.member")'

    print '### 登録済みのチェーンコードを確認'
    cmd_params = {
        'option': e_options_str,
        'container': client_name,
        'channel': config['fabric']['channel'],
    }
    cmd = 'docker exec {option} {container} peer chaincode list --instantiated -C {channel}'.format(**cmd_params)
    result = subprocess.check_output(cmd, shell=True)
    lines = result.splitlines()[1:]
    instantiated = [line for line in lines if 'Name: {},'.format(chaincode_name) in line]
    print 'チェーンコード「{}」の登録済み数: {}'.format(chaincode_name, len(instantiated))

    print '### チェーンコードをインストール'
    cmd_params = {
        'option': e_options_str,
        'container': client_name,
        'name': chaincode_name,
        'version': version,
        'path': chaincode_path,
    }
    cmd = 'docker exec {option} {container} peer chaincode install ' \
          '-n {name} -v {version} -p {path}'.format(**cmd_params)
    print cmd
    subprocess.check_call(cmd, shell=True)

    msg = 'チェーンコードを初期設定'
    command = 'instantiate'

    if len(instantiated) > 0:
        msg = 'チェーンコードをアップグレード'
        command = 'upgrade'

    print '### ' + msg
    cmd_params = {
        'option': e_options_str,
        'container': client_name,
        'command': command,
        'orderer': config['fabric']['orderer']['host'] + ':' + str(config['fabric']['orderer']['port']),
        'channel': config['fabric']['channel'],
        'name': chaincode_name,
        'version': version,
        'construct': chaincode_initialize_args.replace('"', r'\"'),
        'policy': policy.replace('"', r'\"'),
    }
    cmd = 'docker exec {option} {container} peer chaincode {command} ' \
        '-o {orderer} -C {channel} -n {name} -v {version} -c "{construct}" -P "{policy}"'.format(**cmd_params)
    print cmd
    subprocess.check_call(cmd, shell=True)


def ask_target():
    # クライアントを対話形式で選択
    while True:
        print 'チェーンコードをデプロイしたいクライアントの番号を選択してください。'
        for i, c in enumerate(clients):
            print '[{}] {}'.format(i + 1, c)
        num = int(raw_input('> ')) - 1
        if 0 <= num < len(clients):
            break
        print '[Error] 番号が正しくありません。'

    target_client = clients[num]

    # chaincodeを対話形式で入力
    print 'デプロイしたいチェーンコードの名前を入力してください。'
    chaincode_name = (raw_input('> ')).strip()

    return target_client, chaincode_name


if __name__ == '__main__':
    args = sys.argv

    # help指定ならusageを表示する
    if len(args) > 1 and (args[1] == '--help' or args[1] == '-h'):
        show_usage()
        exit()

    if len(args) == 1:
        # 指定がなければ選択式で起動する
        cli, chaincode = ask_target()

    else:
        # 指定があればそれを利用する
        # 引数の数チェック
        if len(args) > 3:
            print '[Error] 引数の数が正しくありません'
            show_usage()
            exit()

        # クライアント設定の確認
        cli = args[1]
        if cli not in clients:
            print '[Error] クライアント「{}」はconfigに記述がありません！'.format(cli)
            exit()

        chaincode = args[2]

    # クライアントの起動確認と起動
    if not check_client(cli):
        print 'クライアント「{}」が起動していません。'.format(cli)
        yes = raw_input('起動しますか？ [y] > ')
        if yes != 'y':
            exit()
        print '### コンテナ{}を起動'.format(cli)
        up.up_container(cli)

    # deployする
    deploy_chaincode(cli, chaincode)
