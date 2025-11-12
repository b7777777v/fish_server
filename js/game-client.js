document.addEventListener('DOMContentLoaded', () => {
    // --- DOM 元素 ---
    const playerIdInput = document.getElementById('playerIdInput');
    const connectBtn = document.getElementById('connectBtn');
    const disconnectBtn = document.getElementById('disconnectBtn');
    const statusSpan = document.getElementById('status');
    const logDiv = document.getElementById('log');
    const actionsDiv = document.getElementById('actions');

    // --- 統計元素 ---
    const messagesSentSpan = document.getElementById('messagesSent');
    const messagesReceivedSpan = document.getElementById('messagesReceived');
    const currentRoomSpan = document.getElementById('currentRoom');
    const fishCountSpan = document.getElementById('fishCount');
    const bulletCountSpan = document.getElementById('bulletCount');
    const latencySpan = document.getElementById('latency');
    const debugInfoDiv = document.getElementById('debugInfo');
    const debugTextSpan = document.getElementById('debugText');

    // --- 新增：玩家信息面板元素 ---
    const playerInfoPanel = document.getElementById('playerInfoPanel');
    const playerNickname = document.getElementById('playerNickname');
    const playerLevel = document.getElementById('playerLevel');
    const playerBalance = document.getElementById('playerBalance');
    const playerExp = document.getElementById('playerExp');
    const refreshPlayerInfoBtn = document.getElementById('refreshPlayerInfoBtn');

    // --- 新增：房間列表面板元素 ---
    const roomListPanel = document.getElementById('roomListPanel');
    const roomListContainer = document.getElementById('roomListContainer');

    // --- 新增：砲台選擇器面板元素 ---
    const cannonSelectorPanel = document.getElementById('cannonSelectorPanel');
    const cannonTypeSelect = document.getElementById('cannonTypeSelect');
    const cannonLevelSelect = document.getElementById('cannonLevelSelect');
    const cannonPowerSlider = document.getElementById('cannonPowerSlider');
    const cannonPowerValue = document.getElementById('cannonPowerValue');
    const applyCannonBtn = document.getElementById('applyCannonBtn');

    // --- 新增：座位信息元素 ---
    const seatsContainer = document.getElementById('seatsContainer');

    // --- 按鈕 ---
    const getRoomListBtn = document.getElementById('getRoomListBtn');
    const joinRoomBtn = document.getElementById('joinRoomBtn');
    const getPlayerInfoBtn = document.getElementById('getPlayerInfoBtn');
    const fireBulletBtn = document.getElementById('fireBulletBtn');
    const switchCannonBtn = document.getElementById('switchCannonBtn');
    const leaveRoomBtn = document.getElementById('leaveRoomBtn');
    const clearLogBtn = document.getElementById('clearLogBtn');
    
    // --- 統計數據 ---
    let stats = {
        messagesSent: 0,
        messagesReceived: 0,
        currentRoom: '',
        fishCount: 0,
        bulletCount: 0,
        latencies: [],
        lastUpdate: null,
        lastFormationCount: 0,
        emptyWarningShown: false
    };

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
        
        // 更新接收統計
        if (type === 'received') {
            stats.messagesReceived++;
            updateStats();
        }
    }
    
    // --- 統計更新功能 ---
    function updateStats() {
        messagesSentSpan.textContent = stats.messagesSent;
        messagesReceivedSpan.textContent = stats.messagesReceived;
        currentRoomSpan.textContent = stats.currentRoom || '無';
        fishCountSpan.textContent = stats.fishCount;
        bulletCountSpan.textContent = stats.bulletCount;
        
        // 計算平均延遲
        if (stats.latencies.length > 0) {
            const avgLatency = stats.latencies.reduce((a, b) => a + b, 0) / stats.latencies.length;
            latencySpan.textContent = Math.round(avgLatency);
        } else {
            latencySpan.textContent = '-';
        }
        
        // 更新調試信息
        updateDebugInfo();
    }
    
    function updateDebugInfo() {
        const info = [
            `已發送: ${stats.messagesSent} 消息`,
            `已接收: ${stats.messagesReceived} 消息`,
            `當前房間: ${stats.currentRoom || '無'}`,
            `遊戲對象: ${stats.fishCount} 魚 + ${stats.bulletCount} 子彈`,
            `最後更新: ${stats.lastUpdate ? stats.lastUpdate.toLocaleTimeString() : '無'}`
        ];
        debugTextSpan.innerHTML = info.join('<br>');
        debugInfoDiv.style.display = 'block';
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

            // 顯示遊戲畫面
            const gameContainer = document.getElementById('gameContainer');
            if (gameContainer) {
                gameContainer.style.display = 'block';
            }

            // 顯示新增的功能面板
            if (playerInfoPanel) playerInfoPanel.style.display = 'block';
            if (roomListPanel) roomListPanel.style.display = 'block';
            if (cannonSelectorPanel) cannonSelectorPanel.style.display = 'block';

            // 啟動遊戲渲染器
            if (window.gameRenderer) {
                // 設置當前玩家
                const currentPlayerId = playerIdInput.value;
                gameRenderer.setCurrentPlayer(currentPlayerId);

                // 添加當前玩家到渲染器
                gameRenderer.addPlayer(currentPlayerId);

                gameRenderer.start();
            }

            // 自動獲取玩家資訊
            setTimeout(() => {
                getPlayerInfoBtn.click();
            }, 500);

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
                // 檢查接收到的數據類型和大小
                if (event.data instanceof ArrayBuffer) {
                    const byteLength = event.data.byteLength;
                    log(`📨 接收到 ${byteLength} 字節的二進位數據`, 'system');
                    
                    if (byteLength === 0) {
                        log('⚠️ 接收到空消息', 'error');
                        return;
                    }
                    
                    // 反序列化 Protobuf
                    const gameMessage = proto.v1.GameMessage.deserializeBinary(event.data);
                    const messageType = gameMessage.getType();
                    // Get message type name properly
                    const messageTypeName = Object.keys(proto.v1.MessageType).find(key => proto.v1.MessageType[key] === messageType) || 'unknown';
                    log(`接收 (S -> C): Type=${messageType} (${messageTypeName}), Size=${byteLength}字節`, 'received');
                    handleServerMessage(gameMessage);
                } else {
                    log(`⚠️ 接收到非預期的數據類型: ${typeof event.data}`, 'error');
                    log(`數據內容: ${event.data}`, 'error');
                }
            } catch (error) {
                log(`❌ 解析 Protobuf 消息時出錯: ${error}`, 'error');
                if (event.data instanceof ArrayBuffer) {
                    const bytes = new Uint8Array(event.data);
                    log(`原始數據 (前50字節): ${Array.from(bytes.slice(0, 50)).map(b => b.toString(16).padStart(2, '0')).join(' ')}`, 'error');
                } else {
                    log(`原始數據: ${event.data}`, 'error');
                }
            }
        };

        socket.onclose = (event) => {
            log(`連接已關閉。 Code: ${event.code}, Reason: ${event.reason}`, 'system');
            statusSpan.textContent = '未連接';
            connectBtn.disabled = false;
            disconnectBtn.disabled = true;
            actionsDiv.style.display = 'none';

            // 隱藏遊戲畫面
            const gameContainer = document.getElementById('gameContainer');
            if (gameContainer) {
                gameContainer.style.display = 'none';
            }

            // 隱藏功能面板
            if (playerInfoPanel) playerInfoPanel.style.display = 'none';
            if (roomListPanel) roomListPanel.style.display = 'none';
            if (cannonSelectorPanel) cannonSelectorPanel.style.display = 'none';

            // 停止遊戲渲染器
            if (window.gameRenderer) {
                gameRenderer.stop();
                gameRenderer.clear();
            }

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
        
        // 更新發送統計
        stats.messagesSent++;
        updateStats();
        
        // Get message type name properly
        const messageTypeName = Object.keys(proto.v1.MessageType).find(key => proto.v1.MessageType[key] === gameMessage.getType()) || 'unknown';
        log(`📤 發送 (C -> S): ${messageTypeName} (Type=${gameMessage.getType()}), Size=${bytes.length}字節`, 'sent');
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
                log(`收到房間列表: ${roomListResp.getRoomsList().length} 個房間`);
                displayRoomList(roomListResp.getRoomsList());
                break;
            case MessageType.JOIN_ROOM_RESPONSE:
                const joinRoomResp = gameMessage.getJoinRoomResponse();
                if (joinRoomResp.getSuccess()) {
                    stats.currentRoom = joinRoomResp.getRoomId();
                    updateStats();
                    log(`✅ 成功加入房間 ${joinRoomResp.getRoomId()}，當前人數: ${joinRoomResp.getPlayerCount()}`);
                } else {
                    log(`❌ 加入房間失敗`, 'error');
                }
                break;
            case MessageType.PLAYER_JOINED:
                const playerJoined = gameMessage.getPlayerJoined();
                const joinedPlayerId = playerJoined.getPlayerId();
                log(`玩家 ${joinedPlayerId} 加入了房間 ${playerJoined.getRoomId()}。`);

                // 添加玩家到渲染器
                if (window.gameRenderer && gameRenderer.isRunning) {
                    gameRenderer.addPlayer(joinedPlayerId);
                }
                break;
            case MessageType.BULLET_FIRED:
                const bulletFired = gameMessage.getBulletFired();
                const bulletPos = bulletFired.getPosition();
                log(`💥 玩家 ${bulletFired.getPlayerId()} 開火了，子彈ID: ${bulletFired.getBulletId()}, 位置: (${bulletPos.getX().toFixed(1)}, ${bulletPos.getY().toFixed(1)})`);
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
                const leftPlayerId = playerLeft.getPlayerId();
                log(`玩家 ${leftPlayerId} 離開了房間。`);

                // 從渲染器移除玩家
                if (window.gameRenderer && gameRenderer.isRunning) {
                    gameRenderer.removePlayer(leftPlayerId);
                }
                break;
            case MessageType.HEARTBEAT_RESPONSE:
                // 心跳回應通常不需要特別處理，但可以記錄
                log(`收到心跳回應。`);
                break;
            case MessageType.SWITCH_CANNON_RESPONSE:
                const switchCannonResp = gameMessage.getSwitchCannonResponse();
                if (switchCannonResp.getSuccess()) {
                    const cannonType = switchCannonResp.getCannonType();
                    const level = switchCannonResp.getLevel();
                    log(`🔧 成功切換砲台類型: ${cannonType}, 等級: ${level}, 威力: ${switchCannonResp.getPower()}`);

                    // 更新渲染器中的砲台
                    if (window.gameRenderer && gameRenderer.isRunning) {
                        const currentPlayerId = playerIdInput.value;
                        gameRenderer.updateCannonType(currentPlayerId, cannonType, level);
                    }
                } else {
                    log(`❌ 切換砲台失敗`, 'error');
                }
                break;
            case MessageType.FIRE_BULLET_RESPONSE:
                const fireBulletResp = gameMessage.getFireBulletResponse();
                if (fireBulletResp.getSuccess()) {
                    log(`💥 成功開火！子彈ID: ${fireBulletResp.getBulletId()}, 消耗: ${fireBulletResp.getCost()}`);
                    console.log('[Client] Fire bullet response received, waiting for ROOM_STATE_UPDATE to show bullet...');
                } else {
                    log(`❌ 開火失敗`, 'error');
                }
                break;
            case MessageType.LEAVE_ROOM_RESPONSE:
                const leaveRoomResp = gameMessage.getLeaveRoomResponse();
                if (leaveRoomResp.getSuccess()) {
                    stats.currentRoom = '';
                    stats.fishCount = 0;
                    stats.bulletCount = 0;
                    updateStats();
                    log(`🚪 成功離開房間 ${leaveRoomResp.getRoomId()}`);
                } else {
                    log(`❌ 離開房間失敗`, 'error');
                }
                break;
            case MessageType.ERROR:
                const errorMsg = gameMessage.getError();
                if (errorMsg) {
                    const errorMessage = errorMsg.getMessage();
                    const errorCode = errorMsg.getCode();
                    log(`❌ 伺服器錯誤 [${errorCode}]: ${errorMessage}`, 'error');
                    
                    // 特殊處理超時錯誤
                    if (errorMessage.includes('timeout')) {
                        log(`⏰ 處理超時 - 伺服器可能過載，請稍後重試`, 'error');
                    }
                } else {
                    log(`❌ 收到未知錯誤消息`, 'error');
                }
                break;
            case MessageType.GET_PLAYER_INFO_RESPONSE:
                const playerInfoResp = gameMessage.getPlayerInfoResponse();
                log(`收到玩家資訊: ${playerInfoResp.getNickname()}, 餘額: ${playerInfoResp.getBalance()}`);
                updatePlayerInfo(playerInfoResp);
                break;
            case MessageType.ROOM_STATE_UPDATE:
                const roomStateUpdate = gameMessage.getRoomStateUpdate();
                handleRoomStateUpdate(roomStateUpdate);
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
        // 獲取當前玩家的砲台信息
        const currentPlayerId = playerIdInput.value;
        let cannonPosition = null;
        let cannonAngle = -Math.PI / 2; // 默認向上

        if (window.gameRenderer && gameRenderer.players.has(currentPlayerId)) {
            // 使用渲染器的統一方法獲取砲口位置
            const barrelEnd = gameRenderer.getBarrelEndPosition(currentPlayerId);
            if (barrelEnd) {
                cannonPosition = { x: barrelEnd.x, y: barrelEnd.y };
                cannonAngle = barrelEnd.angle;

                // 只在開火時記錄，不是每次都記錄
                if (stats.messagesSent % 10 === 0) { // 每10次記錄一次
                    log(`🎯 從砲口發射: 位置(${cannonPosition.x.toFixed(1)}, ${cannonPosition.y.toFixed(1)}), 角度=${(cannonAngle * 180 / Math.PI).toFixed(1)}°, 砲管長=${barrelEnd.barrelLength}`, 'system');
                }
            } else {
                cannonPosition = { x: 600, y: 750 };
                log(`⚠️ 無法獲取砲台位置`, 'error');
            }
        } else {
            // 如果渲染器沒有運行，使用默認位置（畫布底部中央）
            cannonPosition = { x: 600, y: 750 };
            log(`⚠️ 使用默認砲台位置`, 'system');
        }

        const gameMessage = new proto.v1.GameMessage();
        gameMessage.setType(MessageType.FIRE_BULLET);
        const fireBulletReq = new proto.v1.FireBulletRequest();
        fireBulletReq.setDirection(cannonAngle);
        fireBulletReq.setPower(50); // 固定威力
        const position = new proto.v1.Position();
        position.setX(cannonPosition.x);
        position.setY(cannonPosition.y);
        fireBulletReq.setPosition(position);
        gameMessage.setFireBullet(fireBulletReq);
        sendMessage(gameMessage);
    });


    leaveRoomBtn.addEventListener('click', () => {
        const gameMessage = new proto.v1.GameMessage();
        gameMessage.setType(MessageType.LEAVE_ROOM);
        gameMessage.setLeaveRoom(new proto.v1.LeaveRoomRequest()); // payload 是空的
        sendMessage(gameMessage);
    });

    clearLogBtn.addEventListener('click', () => {
        logDiv.innerHTML = '';
        stats = {
            messagesSent: 0,
            messagesReceived: 0,
            currentRoom: stats.currentRoom, // 保留當前房間
            fishCount: 0,
            bulletCount: 0,
            latencies: [],
            lastUpdate: null
        };
        updateStats();
        log('日誌已清除', 'system');
    });

    // 測試渲染器按鈕
    const testRenderBtn = document.getElementById('testRenderBtn');
    if (testRenderBtn) {
        testRenderBtn.addEventListener('click', () => {
            if (window.gameRenderer) {
                if (!gameRenderer.isRunning) {
                    gameRenderer.start();
                    const gameContainer = document.getElementById('gameContainer');
                    if (gameContainer) {
                        gameContainer.style.display = 'block';
                    }
                }
                gameRenderer.addTestData();
                log('🧪 已添加測試數據到渲染器', 'system');
            } else {
                log('❌ 渲染器未初始化', 'error');
            }
        });
    }

    /**
     * 處理房間狀態更新，顯示詳細的遊戲渲染信息
     * @param {proto.v1.RoomStateUpdate} roomStateUpdate - 房間狀態更新消息
     */
    function handleRoomStateUpdate(roomStateUpdate) {
        const fishCount = roomStateUpdate.getFishesList().length;
        const bulletCount = roomStateUpdate.getBulletsList().length;
        const playerCount = roomStateUpdate.getPlayerCount();
        const roomStatus = roomStateUpdate.getRoomStatus();
        const timestamp = roomStateUpdate.getTimestamp();

        // 更新統計
        stats.fishCount = fishCount;
        stats.bulletCount = bulletCount;
        stats.lastUpdate = new Date();

        // 更新座位信息
        const seats = roomStateUpdate.getSeatsList();
        if (seats && seats.length > 0) {
            updateSeatsInfo(seats);
        }

        // 計算延遲
        const now = Date.now();
        const serverTime = timestamp * 1000;
        const latency = now - serverTime;
        stats.latencies.push(latency);
        if (stats.latencies.length > 10) {
            stats.latencies.shift(); // 只保留最近10次的延遲
        }
        updateStats();

        // 更新遊戲渲染器
        if (window.gameRenderer) {
            if (gameRenderer.isRunning) {
                gameRenderer.updateGameState(roomStateUpdate);
                // 減少日誌頻率 - 只在有子彈變化時記錄
                if (bulletCount !== stats.bulletCount) {
                    console.log(`[Client] Passed state to renderer: ${fishCount} fish, ${bulletCount} bullets`);
                }
            } else {
                console.warn('[Client] Renderer exists but is not running!');
            }
        } else {
            console.error('[Client] gameRenderer not found in window object!');
        }

        // 基本狀態信息 - 減少日誌頻率
        if (fishCount > 0 || bulletCount > 0) {
            log(`🎮 房間狀態更新: ${fishCount} 條魚, ${bulletCount} 發子彈, ${playerCount} 位玩家 [${roomStatus}] 延遲:${latency}ms`);
        }

        // 詳細魚類信息（前端渲染需要的數據）- 減少日誌
        if (fishCount > 0 && fishCount !== stats.fishCount) {
            log(`🐟 魚類數量: ${fishCount} 條`);
        }

        // 詳細子彈信息（前端渲染需要的數據）- 減少日誌
        if (bulletCount > 0 && bulletCount !== stats.bulletCount) {
            log(`💥 子彈數量: ${bulletCount} 發`);
        }

        // 魚群陣型信息 - 只在有陣型時顯示
        const formations = roomStateUpdate.getFormationsList();
        if (formations && formations.length > 0 && formations.length !== stats.lastFormationCount) {
            log(`🎯 魚群陣型: ${formations.length} 個陣型`);
            stats.lastFormationCount = formations.length;
        }

        // 如果沒有魚類和子彈，提示可能的問題（只提示一次）
        if (fishCount === 0 && bulletCount === 0 && !stats.emptyWarningShown) {
            log(`⚠️ 注意: 沒有魚類和子彈數據 - 檢查遊戲是否正常運行或房間是否為空`, 'error');
            stats.emptyWarningShown = true;
        } else if (fishCount > 0 || bulletCount > 0) {
            stats.emptyWarningShown = false;
        }
    }

    /**
     * 更新玩家信息面板
     * @param {proto.v1.PlayerInfoResponse} playerInfo - 玩家信息
     */
    function updatePlayerInfo(playerInfo) {
        if (playerNickname) playerNickname.textContent = playerInfo.getNickname() || '-';
        if (playerLevel) playerLevel.textContent = playerInfo.getLevel() || '-';
        if (playerBalance) playerBalance.textContent = playerInfo.getBalance() || '0';
        if (playerExp) playerExp.textContent = playerInfo.getExp() || '0';
    }

    /**
     * 顯示房間列表
     * @param {Array} rooms - 房間列表
     */
    function displayRoomList(rooms) {
        if (!roomListContainer) return;

        if (rooms.length === 0) {
            roomListContainer.innerHTML = '<p style="color: #888;">目前沒有可用的房間</p>';
            return;
        }

        let html = '<div style="display: flex; flex-direction: column; gap: 10px;">';
        rooms.forEach(room => {
            const roomId = room.getRoomId();
            const roomName = room.getName();
            const roomType = room.getType();
            const playerCount = room.getPlayerCount();
            const maxPlayers = room.getMaxPlayers();
            const status = room.getStatus();

            const isFull = playerCount >= maxPlayers;
            const statusColor = status === 'playing' ? '#28a745' : status === 'waiting' ? '#ffc107' : '#6c757d';
            const statusText = status === 'playing' ? '遊戲中' : status === 'waiting' ? '等待中' : '關閉';

            html += `
                <div style="background: white; padding: 10px; border-radius: 5px; border: 1px solid #ddd; cursor: ${isFull ? 'not-allowed' : 'pointer'}; opacity: ${isFull ? '0.6' : '1'};"
                     onclick="${isFull ? '' : `window.joinRoomById('${roomId}')`}">
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                        <div>
                            <strong>${roomName}</strong>
                            <span style="background: #e9ecef; padding: 2px 6px; border-radius: 3px; font-size: 12px; margin-left: 8px;">${roomType}</span>
                        </div>
                        <div style="text-align: right;">
                            <span style="color: ${statusColor}; font-weight: bold;">${statusText}</span>
                            <div style="font-size: 12px; color: #666;">👥 ${playerCount}/${maxPlayers}</div>
                        </div>
                    </div>
                    ${isFull ? '<div style="color: #dc3545; font-size: 12px; margin-top: 5px;">房間已滿</div>' : ''}
                </div>
            `;
        });
        html += '</div>';
        roomListContainer.innerHTML = html;
    }

    /**
     * 加入指定房間
     * @param {string} roomId - 房間ID
     */
    window.joinRoomById = function(roomId) {
        const gameMessage = new proto.v1.GameMessage();
        gameMessage.setType(MessageType.JOIN_ROOM);
        const joinRoomReq = new proto.v1.JoinRoomRequest();
        joinRoomReq.setRoomId(roomId);
        gameMessage.setJoinRoom(joinRoomReq);
        sendMessage(gameMessage);
        log(`正在加入房間 ${roomId}...`, 'system');
    };

    /**
     * 更新座位信息顯示
     * @param {Array} seats - 座位列表
     */
    function updateSeatsInfo(seats) {
        if (!seatsContainer) return;

        const currentPlayerId = playerIdInput.value;
        let html = '';

        seats.forEach(seat => {
            const seatId = seat.getSeatId();
            const playerId = seat.getPlayerId();
            const nickname = seat.getNickname();

            const isEmpty = !playerId || playerId === '0';
            const isCurrentPlayer = playerId === currentPlayerId;
            const seatColor = isCurrentPlayer ? '#28a745' : isEmpty ? '#6c757d' : '#007bff';
            const seatIcon = isEmpty ? '🪑' : isCurrentPlayer ? '⭐' : '👤';

            html += `
                <div style="margin-bottom: 3px; padding: 3px 6px; background: ${isEmpty ? 'rgba(108,117,125,0.1)' : isCurrentPlayer ? 'rgba(40,167,69,0.2)' : 'rgba(0,123,255,0.1)'}; border-radius: 3px; display: flex; justify-content: space-between;">
                    <span>${seatIcon} 座位 ${seatId + 1}</span>
                    <span style="color: ${seatColor}; font-weight: ${isCurrentPlayer ? 'bold' : 'normal'};">
                        ${isEmpty ? '空位' : nickname || `玩家${playerId}`}
                    </span>
                </div>
            `;
        });

        seatsContainer.innerHTML = html || '<div style="color: #888;">無座位資訊</div>';
    }

    // --- 新增功能的事件監聽器 ---

    // 刷新玩家資訊
    if (refreshPlayerInfoBtn) {
        refreshPlayerInfoBtn.addEventListener('click', () => {
            getPlayerInfoBtn.click();
        });
    }

    // 砲台威力滑桿
    if (cannonPowerSlider) {
        cannonPowerSlider.addEventListener('input', (e) => {
            if (cannonPowerValue) {
                cannonPowerValue.textContent = e.target.value;
            }
        });
    }

    // 應用砲台設置
    if (applyCannonBtn) {
        applyCannonBtn.addEventListener('click', () => {
            const cannonType = parseInt(cannonTypeSelect.value);
            const cannonLevel = parseInt(cannonLevelSelect.value);
            const power = parseInt(cannonPowerSlider.value);

            const gameMessage = new proto.v1.GameMessage();
            gameMessage.setType(MessageType.SWITCH_CANNON);
            const switchCannonReq = new proto.v1.SwitchCannonRequest();
            switchCannonReq.setCannonType(cannonType);
            switchCannonReq.setLevel(cannonLevel);
            gameMessage.setSwitchCannon(switchCannonReq);
            sendMessage(gameMessage);

            log(`🔧 切換砲台: Type ${cannonType}, Level ${cannonLevel}, Power ${power}`, 'system');
        });
    }

    // 初始化統計顯示
    updateStats();
    log('🚀 遊戲客戶端已載入，準備連接...', 'system');
});
