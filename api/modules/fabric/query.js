'use strict';
/**
 * @fileoverview chaincodeに対してqueryの送信を行うモジュール
 */

// Chaincode Query
const FabricClient = require('fabric-client');
const config = require('config');
const path = require('path');

const fabricConfig = config.get('fabric');

const channelName = fabricConfig['channel'];
const peerGrpcAddressList = fabricConfig['peers'];
const storePath = path.join(__dirname, '../../../hfc-key-store');
const userName = fabricConfig['user'];

class FabricQuery {
    constructor(chaincode) {
        this.fabricClient = new FabricClient();

        this.channel = this.fabricClient.newChannel(channelName);
        const peer = this.fabricClient.newPeer(peerGrpcAddressList[0]);
        this.channel.addPeer(peer);
        this.chaincode = chaincode;
        this.memberUser = null;
        this.storeConf = {
            path: storePath
        };
    }

    getUserContext(stateStore) {
        this.fabricClient.setStateStore(stateStore);
        const cryptoSuite = FabricClient.newCryptoSuite();

        const cryptoStore = FabricClient.newCryptoKeyStore(this.storeConf);
        cryptoSuite.setCryptoKeyStore(cryptoStore);
        this.fabricClient.setCryptoSuite(cryptoSuite);

        return this.fabricClient.getUserContext(userName, true);
    }

    sendQueryProposal(userFromStore, request) {
        if (userFromStore && userFromStore.isEnrolled()) {
            this.memberUser = userFromStore;
        } else {
            throw new Error('Failed: ' + userName);
        }

        return this.channel.queryByChaincode(request);
    }

    parseResult(queryResponses) {
        if (queryResponses && queryResponses.length > 0) {
            return JSON.parse(queryResponses[0].toString());
        }

        return {'error': 'No payloads were returned'};
    }

    async run(functionName, args) {
        const stateStore = await FabricClient.newDefaultKeyValueStore(this.storeConf);
        const userFromStore = await this.getUserContext(stateStore);
        const request = {
            chaincodeId: this.chaincode,
            fcn: functionName,
            args: args
        };
        const queryResult = await this.sendQueryProposal(userFromStore, request);
        const result = this.parseResult(queryResult);

        return result;
    }
}

module.exports = FabricQuery;
