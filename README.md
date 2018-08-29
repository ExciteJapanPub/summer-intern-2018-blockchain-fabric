# summer-intern-2018-blockchain-fabric
サマーインターン2018ブロックチェーン回のFabricネットワーク構築スクリプトおよびAPI

## Requirement

- HyperLedgerFabric v1.2
- python 2.7.x
- go 1.9.x
- node 8.11.x
- npm 5.6.x

## Install
```
pip install pyyaml  # -> 未インストールなら
./bin/get-docker-image.sh # -> Docker Imageがpullされてなければ
npm install
npm run linklocal
```

## Fabricネットワーク用スクリプト

指定する引数などは「-h」で確認可

### clean.py

コンテナやネットワークなどすべてを削除

`python script/clean.py`

### up.py

コンテナを起動

`python script/up.py`

### setup_channel.py

チャンネルを作成してpeerを登録

`python script/setup_channel.py`

※既存のチャンネルにpeerを登録する機能は載せていない

### deploy_chaincode.py

チェーンコードをinstall/instantiate/upgrade

バージョン名は動的に生成され、instantiateとupgradeのどちらを使うべきかは自動で判断される。

`python script/deploy_chaincode.py`

## npm scripts

### adminとuserの登録

`npm run register`

[fabric-samples](https://github.com/hyperledger/fabric-samples/tree/release-1.1/fabcar)のものをそのまま利用。

### ローカルモジュールへのシンボリックリンク作成

`npm run linklocal`

## API

本来であれば、  
チェーンコード毎に用途別でエンドポイントを作成して、  
それぞれバリデーション等々を行うべきですが、  
今回に関しては期間の都合やテーマの関係でAPI部分に関しては、  
触れない想定なので指定のchaincodeにメソッドと引数を渡すだけのもの用意しています。

### 起動

```
node api/app.js
```

### Document

https://github.com/pages/ExciteJapanPub/summer-intern-2018-blockchain-fabric


## サンプルチェーンコードに関して

chaincode/以下に公式で提供されているfabcarのほかにいくつかサンプルを用意したので  
簡単な説明を載せておきます。

* entry (勤怠など日毎になんらかのデータを登録して行くサンプル)
* point (簡単なポイント管理)
* rental (備品の貸し出し管理)
* smartLock (スマートロックの施錠管理)
* supply (在庫および配送状況管理)

簡単化のため必要最低限のメソッドと処理しか実装していないので、  
不足を感じた部分があれば実際にチェーンコードを作成する際に追加してみてください
