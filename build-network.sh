#!/bin/bash
#
## 配置基本环境变量，启动网络
################################################################################
#
# fabric主目录
export FABRIC_SAMPLE_PATH=/root/go/src/github.com/hyperledger/fabric-samples
# 网络目录
export NETWORK_PATH=${FABRIC_SAMPLE_PATH}/test-network
# 通道名称
export CHANNEL_NAME="userInfo"
# 链码目录
export CC_SRC_PATH=/root/go/src/github.com/hiro942/sekiro/chaincode
# 链码包的名称
export CC_NAME="basic"
# fabric配置目录，包含fabric-sample仓库的core.yaml
export FABRIC_CFG_PATH=${FABRIC_SAMPLE_PATH}/config
# 各节点的网络地址address
export ORDERER_ADDRESS=localhost:7050
export PEER0_ORG1_ADDRESS=localhost:7051
export PEER0_ORG2_ADDRESS=localhost:9051
# orderer的ca证书
export ORDERER_CA_FILE="${NETWORK_PATH}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem"
export PEER0_ORG1_TLSROOTCERT="${NETWORK_PATH}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
export PEER0_ORG2_TLSROOTCERT="${NETWORK_PATH}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt"


${NETWORK_PATH}/./network.sh down
${NETWORK_PATH}/./network.sh up createChannel -c ${CHANNEL_NAME} -ca
${NETWORK_PATH}/./network.sh deployCC -ccn ${CC_NAME} -ccp ${CC_SRC_PATH} -ccl golang
./network.sh deployCC -ccn basic -ccp /root/go/src/github.com/hiro942/sekiro/blockchain/chaincode -ccl go
go run /root/go/src/github.com/hiro942/sekiro/main.go
## 关闭并重新启动网，创建一个userInfo通道，并使用CA生成网络成员的加密材料
#${NETWORK_PATH}/./network.sh down
#${NETWORK_PATH}/./network.sh up createChannel -c ${CHANNEL_NAME} -ca
#
#
## 使用peer cli进行链码的安装和部署
## 将peer cli配置进环境变量
#export PATH=${PATH}:/root/go/src/github.com/hyperledger/fabric-samples/bin
#
## 1.打包 - 将链码打包为一压缩文件
################################################################################
#
#peer lifecycle chaincode package ${CC_NAME}.tar.gz --path ${CC_SRC_PATH} --lang golang --label ${CC_NAME}_1.0
#
## 2.安装 - 在peer节点上安装链码
################################################################################
#
## 2.1 为组织1安装
## 配置环境变量，以组织1 Admin user 的身份使用 peer cli
#export CORE_PEER_TLS_ENABLED=true
#export CORE_PEER_LOCALMSPID="Org1MSP"
#export CORE_PEER_TLS_ROOTCERT_FILE=${NETWORK_PATH}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
#export CORE_PEER_MSPCONFIGPATH=${NETWORK_PATH}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
#export CORE_PEER_ADDRESS=${PEER0_ORG1_ADDRESS}
#peer lifecycle chaincode install ${CC_NAME}.tar.gz
#peer lifecycle chaincode queryinstalled
#
## 2.2 为组织2安装
#export CORE_PEER_LOCALMSPID="Org2MSP"
#export CORE_PEER_TLS_ROOTCERT_FILE=${NETWORK_PATH}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
#export CORE_PEER_MSPCONFIGPATH=${NETWORK_PATH}/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
#export CORE_PEER_ADDRESS=${PEER0_ORG2_ADDRESS}
#peer lifecycle chaincode install ${CC_NAME}.tar.gz
#peer lifecycle chaincode queryinstalled
#
## 3.批准 - 批准链码定义
################################################################################
#
#export CC_PKG_ID=
#
## 此时环境变量指向组织2的Admin，先由组织2批准 (命令中sequence表示链码定义和更新的次数)
#peer lifecycle chaincode approveformyorg -o ${ORDERER_ADDRESS} --ordererTLSHostnameOverride orderer.example.com --channelID ${CHANNEL_NAME} --name ${CC_NAME} --version 1.0 --package-id ${CC_PKG_ID} --sequence 1 --tls --cafile ${ORDERER_CA_FILE}
#
#export CORE_PEER_TLS_ENABLED=true
#export CORE_PEER_LOCALMSPID="Org1MSP"
#export CORE_PEER_TLS_ROOTCERT_FILE=${NETWORK_PATH}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
#export CORE_PEER_MSPCONFIGPATH=${NETWORK_PATH}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
#export CORE_PEER_ADDRESS=${PEER0_ORG1_ADDRESS}
#
#peer lifecycle chaincode approveformyorg -o ${ORDERER_ADDRESS} --ordererTLSHostnameOverride orderer.example.com --channelID ${CHANNEL_NAME} --name ${CC_NAME} --version 1.0 --package-id ${CC_PKG_ID} --sequence 1 --tls --cafile ${ORDERER_CA_FILE}
#
#
## 4.提交 - 由其中一个组织将链码提交至通道
################################################################################
#
## 足够数量的组织批准该链码定义之后，链码就可以提交至通道并与通道账本交互
## 查看是否满足提交条件
#peer lifecycle chaincode checkcommitreadiness --channelID ${CHANNEL_NAME} --name ${CC_NAME} --version 1.0 --sequence 1 --tls --cafile ${ORDERER_CA_FILE} --output json
#
## 提交链码（同样需要由组织的Admin来提交）
#peer lifecycle chaincode commit -o ${ORDERER_ADDRESS} --ordererTLSHostnameOverride orderer.example.com --channelID ${CHANNEL_NAME} --name ${CC_NAME} --version 1.0 --sequence 1 --tls --cafile ${ORDERER_CA_FILE} --peerAddresses ${PEER0_ORG1_ADDRESS} --tlsRootCertFiles ${PEER0_ORG1_TLSROOTCERT} --peerAddresses ${PEER0_ORG2_ADDRESS} --tlsRootCertFiles ${PEER0_ORG2_TLSROOTCERT}
#
## 验证提交
#peer lifecycle chaincode querycommitted --channelID ${CHANNEL_NAME} --name ${CC_NAME} --cafile ${ORDERER_CA_FILE}
#
## 5.调用链码
################################################################################
#
## 调用：InitLeger
#peer chaincode invoke -o ${ORDERER_ADDRESS} --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${ORDERER_CA_FILE} -C ${CHANNEL_NAME} -n ${CC_NAME} --peerAddresses ${PEER0_ORG1_ADDRESS} --tlsRootCertFiles ${PEER0_ORG1_TLSROOTCERT} --peerAddresses ${PEER0_ORG2_ADDRESS} --tlsRootCertFiles ${PEER0_ORG2_TLSROOTCERT} -c '{"function":"InitLedger","Args":[]}'
## 调用：GetAllUser
#peer chaincode invoke -o ${ORDERER_ADDRESS} --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${ORDERER_CA_FILE} -C ${CHANNEL_NAME} -n ${CC_NAME} --peerAddresses ${PEER0_ORG1_ADDRESS} --tlsRootCertFiles ${PEER0_ORG1_TLSROOTCERT} --peerAddresses ${PEER0_ORG2_ADDRESS} --tlsRootCertFiles ${PEER0_ORG2_TLSROOTCERT} -c '{"function":"GetAllUsers","Args":[]}'



