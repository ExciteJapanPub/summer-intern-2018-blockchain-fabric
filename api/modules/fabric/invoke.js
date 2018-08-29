'use strict';
/**
 * @fileoverview chaincodeのinvokeを行うモジュール
 */

const FabricClient = require('fabric-client');
const config = require('config');
const path = require('path');
const httpStatus = require('http-status-codes');

const fabricConfig = config.get('fabric');

const channelName = fabricConfig['channel'];
const peerGrpcAddressList = fabricConfig['peers'];
const orderGrpcAddressList = fabricConfig['orderers'];
const storePath = path.join(__dirname, '../../../hfc-key-store');
const userName = fabricConfig['user'];

class FabricInvoke {
    constructor(chaincode) {
        this.fabricClient = new FabricClient();
        this.chaincode = chaincode;
        this.channel = this.fabricClient.newChannel(channelName);

        const peer = this.fabricClient.newPeer(peerGrpcAddressList[0]);
        this.channel.addPeer(peer);

        const order = this.fabricClient.newOrderer(orderGrpcAddressList[0]);
        this.channel.addOrderer(order);

        this.memberUser = null;
        this.storeConf = {
            path: storePath
        };

        this.txId = null;
    }
    getUserContext(stateStore) {
        this.fabricClient.setStateStore(stateStore);
        const cryptoSuite = FabricClient.newCryptoSuite();

        const cryptoStore = FabricClient.newCryptoKeyStore(this.storeConf);
        cryptoSuite.setCryptoKeyStore(cryptoStore);
        this.fabricClient.setCryptoSuite(cryptoSuite);

        return this.fabricClient.getUserContext(userName, true);
    }

    sendTransactionProposal(userFromStore, request) {
        if (userFromStore && userFromStore.isEnrolled()) {
            this.memberUser = userFromStore;
        } else {
            throw new Error('Failed: ' + userName);
        }

        this.txId = this.fabricClient.newTransactionID();

        request.chainId = channelName;
        request.txId = this.txId;

        return this.channel.sendTransactionProposal(request);
    }

    processTransaction(transcationResult) {
        const proposalResponses = transcationResult[0];
        const proposal = transcationResult[1];
        const isProposalGood = proposalResponses &&
                         proposalResponses[0].response &&
                         proposalResponses[0].response.status === httpStatus.OK;

        if (!isProposalGood) {
            throw new Error('Failed: Transaction proposal');
        }

        const request = {
            proposalResponses: proposalResponses,
            proposal: proposal
        };

        const transactionIdString = this.txId.getTransactionID();
        const promises = [];

        const sendPromise = this.channel.sendTransaction(request);
        promises.push(sendPromise);

        const eventHub = this.fabricClient.newEventHub();
        eventHub.setPeerAddr(peerGrpcAddressList[1]);

        function transactionExecutor(resolve, reject) {
            function onTimeout() {
                eventHub.disconnect();
                resolve({
                    eventStatus: 'TIMEOUT'
                });
            }

            function onEvent(tx, code) {
                clearTimeout(handle);
                eventHub.unregisterTxEvent(transactionIdString);
                eventHub.disconnect();

                const returnStatus = {
                    eventStatus: code,
                    txId: transactionIdString
                };
                if (code !== 'VALID') {
                    resolve(returnStatus);
                } else {
                    resolve(returnStatus);
                }
            }

            function onError(err) {
                reject(new Error('eventhub error :' + err));
            }

            const handle = setTimeout(onTimeout, fabricConfig['timeout']);
            eventHub.connect();
            eventHub.registerTxEvent(transactionIdString, onEvent, onError);
        }

        const txPromise = new Promise(transactionExecutor);
        promises.push(txPromise);

        return Promise.all(promises);
    }

    isSuccess(results) {
        const ordererResult = results && results[0] && results[0].status === 'SUCCESS';
        const commitResult = results && results[1] && results[1].eventStatus === 'VALID';

        return ordererResult && commitResult;
    }

    async run(functionName, args) {
        const stateStore = await FabricClient.newDefaultKeyValueStore(this.storeConf);
        const userFromStore = await this.getUserContext(stateStore);
        const request = {
            chaincodeId: this.chaincode,
            fcn: functionName,
            args: args
        };
        const proposalResult = await this.sendTransactionProposal(userFromStore, request);
        const transcationResult = await this.processTransaction(proposalResult);
        const result = this.isSuccess(transcationResult);

        if (result) {
            return JSON.parse(proposalResult[0][0].response.payload.toString());
        }

        return {'status': httpStatus.INTERNAL_SERVER_ERROR, 'message': 'Failed invoke'};
    }
}

module.exports = FabricInvoke;
