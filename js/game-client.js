document.addEventListener('DOMContentLoaded', () => {
    // --- DOM 元素 ---
    const playerIdInput = document.getElementById('playerIdInput');
    const connectBtn = document.getElementById('connectBtn');
    const disconnectBtn = document.getElementById('disconnectBtn');
    const statusSpan = document.getElementById('status');
    const logDiv = document.getElementById('log');
    const actionsDiv = document.getElementById('actions');

    // --- 按鈕 ---
    const getRoomListBtn = document.getElementById('getRoomListBtn');
    const joinRoomBtn = document.getElementById('joinRoomBtn');
    const getPlayerInfoBtn = document.getElementById('getPlayerInfoBtn');
    const fireBulletBtn = document.getElementById('fireBulletBtn');
    const switchCannonBtn = document.getElementById('switchCannonBtn');
    const leaveRoomBtn = document.getElementById('leaveRoomBtn');

    // --- WebSocket 相關 ---
    const WEBSOCKET_URL = 'ws://localhost:9090/ws';
    let socket = null;
    let heartbeatInterval = null;

    // 直接使用 Protobuf 生成的 MessageType 枚舉
    const MessageType = proto.v1.MessageType;

    // --- 日誌功能 ---
    function log(message, type = 'system') {
        const entry = document.createElement('div');
        entry.className = `log-entry ${type}`;
        entry.textContent = `[${new Date().toLocaleTimeString()}] ${message}`;
        logDiv.appendChild(entry);
        logDiv.scrollTop = logDiv.scrollHeight; // 自動滾動到底部
    }

    // --- WebSocket 核心功能 ---
    function connect() {
        if (socket && socket.readyState === WebSocket.OPEN) {
            log('已經連接。', 'system');
            return;
        }

        const playerId = playerIdInput.value;
        if (!playerId) {
            log('請輸入玩家ID。', 'error');
            return;
        }

        const url = `${WEBSOCKET_URL}?player_id=${encodeURIComponent(playerId)}`;
        log(`正在連接到 ${url}`, 'system');

        socket = new WebSocket(url);
        // 設置 WebSocket 接收二進位數據
        socket.binaryType = "arraybuffer";

        socket.onopen = () => {
            log('成功連接到伺服器', 'system');
            statusSpan.textContent = '已連接';
            connectBtn.disabled = true;
            disconnectBtn.disabled = false;
            actionsDiv.style.display = 'block';

            // 建立心跳機制
            heartbeatInterval = setInterval(() => {
                const heartbeatMsg = new proto.v1.GameMessage();
                heartbeatMsg.setType(MessageType.HEARTBEAT);
                const heartbeatPayload = new proto.v1.HeartbeatMessage();
                heartbeatPayload.setTimestamp(Date.now());
                heartbeatMsg.setHeartbeat(heartbeatPayload);
                sendMessage(heartbeatMsg);
            }, 30000); // 每 30 秒發送一次心跳
        };

        socket.onmessage = (event) => {
            try {
                // 接收到的是 ArrayBuffer，需要反序列化 Protobuf
                const gameMessage = proto.v1.GameMessage.deserializeBinary(event.data);
                log(`接收 (S -> C): Type=${gameMessage.getType()}, Payload=${JSON.stringify(gameMessage.toObject())}`, 'received');
                handleServerMessage(gameMessage);
            } catch (error) {
                log(`解析 Protobuf 消息時出錯: ${error}`, 'error');
                log(`原始數據 (ArrayBuffer): ${new Uint8Array(event.data)}`, 'error');
            }
        };

        socket.onclose = (event) => {
            log(`連接已關閉。 Code: ${event.code}, Reason: ${event.reason}`, 'system');
            statusSpan.textContent = '未連接';
            connectBtn.disabled = false;
            disconnectBtn.disabled = true;
            actionsDiv.style.display = 'none';

            // 清除心跳
            if (heartbeatInterval) {
                clearInterval(heartbeatInterval);
                heartbeatInterval = null;
            }
        };

        socket.onerror = (error) => {
            log('WebSocket 發生錯誤。請檢查伺服器是否正在運行，或查看瀏覽器開發者工具的控制台以獲取詳細資訊。', 'error');
            console.error('WebSocket Error:', error);
        };
    }

    function disconnect() {
        if (socket) {
            socket.close();
        }
    }

    /**
     * 封裝並發送 Protobuf 消息
     * @param {proto.v1.GameMessage} gameMessage - 已經建立好的 Protobuf GameMessage 物件
     */
    function sendMessage(gameMessage) {
        if (!socket || socket.readyState !== WebSocket.OPEN) {
            log('無法發送消息：未連接到伺服器。', 'error');
            return;
        }

        const bytes = gameMessage.serializeBinary();
        socket.send(bytes);
        log(`發送 (C -> S): Type=${gameMessage.getType()}, Payload=${JSON.stringify(gameMessage.toObject())}`, 'sent');
    }

    /**
     * 根據消息類型處理來自伺服器的 Protobuf 消息
     * @param {proto.v1.GameMessage} gameMessage - 從伺服器收到的已解析的 Protobuf GameMessage
     */
    function handleServerMessage(gameMessage) {
        const type = gameMessage.getType();

        switch (type) {
            case MessageType.WELCOME:
                const welcomeMsg = gameMessage.getWelcome();
                if (welcomeMsg) {
                    log(`伺服器歡迎您: ClientID=${welcomeMsg.getClientId()}, ServerTime=${welcomeMsg.getServerTime()}`);
                } else {
                    log('收到 WELCOME 訊息，但缺少 payload。', 'error');
                }
                break;
            case MessageType.ROOM_LIST_RESPONSE:
                const roomListResp = gameMessage.getRoomListResponse();
                log(`收到房間列表: ${JSON.stringify(roomListResp.toObject())}`);
                break;
            case MessageType.JOIN_ROOM_RESPONSE:
                const joinRoomResp = gameMessage.getJoinRoomResponse();
                if (joinRoomResp.getSuccess()) {
                    log(`成功加入房間 ${joinRoomResp.getRoomId()}，當前人數: ${joinRoomResp.getPlayerCount()}`);
                } else {
                    log(`加入房間失敗`, 'error');
                }
                break;
            case MessageType.PLAYER_JOINED:
                const playerJoined = gameMessage.getPlayerJoined();
                log(`玩家 ${playerJoined.getPlayerId()} 加入了房間 ${playerJoined.getRoomId()}。`);
                break;
            case MessageType.BULLET_FIRED:
                const bulletFired = gameMessage.getBulletFired();
                log(`玩家 ${bulletFired.getPlayerId()} 開火了，子彈ID: ${bulletFired.getBulletId()}`);
                break;
            case MessageType.FISH_SPAWNED:
                const fishSpawnedOld = gameMessage.getFishSpawned();
                log(`魚 ${fishSpawnedOld.getFishId()} (類型: ${fishSpawnedOld.getFishType()}) 出現了！`);
                break;
            case MessageType.FISH_DIED:
                const fishDied = gameMessage.getFishDied();
                log(`魚 ${fishDied.getFishId()} 被捕獲！玩家 ${fishDied.getPlayerId()} 獲得獎勵 ${fishDied.getReward()} 金幣。`);
                break;
            case MessageType.PLAYER_REWARD:
                const playerReward = gameMessage.getPlayerReward();
                log(`玩家 ${playerReward.getPlayerId()} 獲得獎勵: ${playerReward.getReward()} 金幣。`);
                break;
            case MessageType.PLAYER_LEFT:
                const playerLeft = gameMessage.getPlayerLeft();
                log(`玩家 ${playerLeft.getPlayerId()} 離開了房間。`);
                break;
            case MessageType.ERROR:
                const errorMsg = gameMessage.getError();
                log(`伺服器錯誤: Code=${errorMsg.getCode()}, Message=${errorMsg.getMessage()}`, 'error');
                break;
            case MessageType.HEARTBEAT_RESPONSE:
                // 心跳回應通常不需要特別處理，但可以記錄
                log(`收到心跳回應。`);
                break;
            case MessageType.SWITCH_CANNON_RESPONSE:
                const switchCannonResp = gameMessage.getSwitchCannonResponse();
                if (switchCannonResp.getSuccess()) {
                    log(`成功切換砲台類型: ${switchCannonResp.getCannonType()}, 等級: ${switchCannonResp.getLevel()}, 威力: ${switchCannonResp.getPower()}`);
                } else {
                    log(`切換砲台失敗`, 'error');
                }
                break;
            case MessageType.GET_PLAYER_INFO_RESPONSE:
                const playerInfoResp = gameMessage.getPlayerInfoResponse();
                log(`收到玩家資訊: ${JSON.stringify(playerInfoResp.toObject())}`);
                break;
            case MessageType.ROOM_STATE_UPDATE:
                const roomStateUpdate = gameMessage.getRoomStateUpdate();
                log(`🎮 房間狀態更新: ${roomStateUpdate.getFishesList().length} 條魚, ${roomStateUpdate.getBulletsList().length} 發子彈`);
                break;
            case MessageType.FISH_SPAWNED:
                const fishSpawnedEvent = gameMessage.getFishSpawned();
                log(`🐟 新魚出現: ID=${fishSpawnedEvent.getFishId()}, 類型=${fishSpawnedEvent.getFishType()}`);
                break;
            // 在這裡添加更多 case 來處理其他消息類型
            default:
                log(`收到未知的 Protobuf 消息類型: ${type}`);
        }
    }

    // --- 綁定事件監聽器 ---
    connectBtn.addEventListener('click', connect);
    disconnectBtn.addEventListener('click', disconnect);

    getRoomListBtn.addEventListener('click', () => {
        const gameMessage = new proto.v1.GameMessage();
        gameMessage.setType(MessageType.GET_ROOM_LIST);
        gameMessage.setGetRoomList(new proto.v1.GetRoomListRequest()); // payload 是空的
        sendMessage(gameMessage);
    });

    joinRoomBtn.addEventListener('click', () => {
        const gameMessage = new proto.v1.GameMessage();
        gameMessage.setType(MessageType.JOIN_ROOM);
        const joinRoomReq = new proto.v1.JoinRoomRequest();
        joinRoomReq.setRoomId("101"); // 假設加入房間 ID 為 "101"
        gameMessage.setJoinRoom(joinRoomReq);
        sendMessage(gameMessage);
    });

    getPlayerInfoBtn.addEventListener('click', () => {
        const gameMessage = new proto.v1.GameMessage();
        gameMessage.setType(MessageType.GET_PLAYER_INFO);
        gameMessage.setGetPlayerInfo(new proto.v1.GetPlayerInfoRequest()); // payload 是空的
        sendMessage(gameMessage);
    });

    fireBulletBtn.addEventListener('click', () => {
        const gameMessage = new proto.v1.GameMessage();
        gameMessage.setType(MessageType.FIRE_BULLET);
        const fireBulletReq = new proto.v1.FireBulletRequest();
        fireBulletReq.setDirection(Math.random() * 2 * Math.PI); // 隨機方向 (弧度)
        fireBulletReq.setPower(Math.floor(Math.random() * 100) + 1); // 隨機威力 1-100
        const position = new proto.v1.Position();
        position.setX(Math.random() * 800); // 隨機 X 位置
        position.setY(Math.random() * 600); // 隨機 Y 位置
        fireBulletReq.setPosition(position);
        gameMessage.setFireBullet(fireBulletReq);
        sendMessage(gameMessage);
    });

    switchCannonBtn.addEventListener('click', () => {
        const gameMessage = new proto.v1.GameMessage();
        gameMessage.setType(MessageType.SWITCH_CANNON);
        const switchCannonReq = new proto.v1.SwitchCannonRequest();
        switchCannonReq.setCannonType(2); // 假設切換到砲台類型為 2
        switchCannonReq.setLevel(1); // 砲台等級為 1
        gameMessage.setSwitchCannon(switchCannonReq);
        sendMessage(gameMessage);
    });

    leaveRoomBtn.addEventListener('click', () => {
        const gameMessage = new proto.v1.GameMessage();
        gameMessage.setType(MessageType.LEAVE_ROOM);
        gameMessage.setLeaveRoom(new proto.v1.LeaveRoomRequest()); // payload 是空的
        sendMessage(gameMessage);
    });
});
