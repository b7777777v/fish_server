/**
 * 遊戲渲染器 - 負責在 Canvas 上繪製遊戲畫面
 */
class GameRenderer {
    constructor(canvasId) {
        this.canvas = document.getElementById(canvasId);
        if (!this.canvas) {
            throw new Error(`Canvas element with id "${canvasId}" not found`);
        }

        this.ctx = this.canvas.getContext('2d');
        this.width = this.canvas.width;
        this.height = this.canvas.height;

        // 遊戲狀態
        this.fishes = [];
        this.bullets = [];
        this.formations = [];

        // 玩家和砲台
        this.players = new Map(); // player_id -> {id, position, cannonType, level}
        this.currentPlayerId = null; // 當前玩家ID

        // 追蹤上次的數量，用於減少日誌
        this.lastFishCount = 0;
        this.lastBulletCount = 0;

        // FPS 追蹤
        this.fps = 0;
        this.frameCount = 0;
        this.lastFpsUpdate = Date.now();

        // 動畫
        this.animationId = null;
        this.isRunning = false;

        // 魚的顏色映射 (根據魚的類型)
        this.fishColors = {
            1: '#FFD700', // 金色 - 小魚
            2: '#FF6B6B', // 紅色 - 中魚
            3: '#4ECDC4', // 青色 - 大魚
            4: '#95E1D3', // 淺綠色
            5: '#F38181', // 粉紅色
            6: '#AA96DA', // 紫色
            7: '#FCBAD3', // 淺粉色
            8: '#FFFFD2', // 淺黃色
        };

        console.log('GameRenderer initialized');
    }

    /**
     * 更新遊戲狀態
     */
    updateGameState(roomStateUpdate) {
        if (!roomStateUpdate) {
            console.warn('updateGameState: roomStateUpdate is null or undefined');
            return;
        }

        // 更新魚類
        this.fishes = roomStateUpdate.getFishesList().map(fish => ({
            id: fish.getFishId(),
            type: fish.getFishType(),
            x: fish.getPosition().getX(),
            y: fish.getPosition().getY(),
            direction: fish.getDirection(),
            speed: fish.getSpeed(),
            health: fish.getHealth(),
            maxHealth: fish.getMaxHealth(),
            value: fish.getValue()
        }));

        // 更新子彈
        const bulletsList = roomStateUpdate.getBulletsList();
        this.bullets = bulletsList.map(bullet => ({
            id: bullet.getBulletId(),
            playerId: bullet.getPlayerId(),
            x: bullet.getPosition().getX(),
            y: bullet.getPosition().getY(),
            direction: bullet.getDirection(),
            speed: bullet.getSpeed(),
            power: bullet.getPower()
        }));

        // 更新魚群陣型
        this.formations = roomStateUpdate.getFormationsList().map(formation => ({
            id: formation.getFormationId(),
            type: formation.getFormationType(),
            centerX: formation.getCenterPosition().getX(),
            centerY: formation.getCenterPosition().getY(),
            progress: formation.getProgress(),
            fishIds: formation.getFishIdsList()
        }));

        // 調試日誌 - 顯示接收到的數據（減少日誌頻率）
        if (this.fishes.length > 0 || this.bullets.length > 0) {
            // 只在對象數量變化時記錄，不是每次都記錄
            const stateChanged = this.fishes.length !== this.lastFishCount ||
                                this.bullets.length !== this.lastBulletCount;
            if (stateChanged) {
                console.log(`[Renderer] Updated: ${this.fishes.length} fishes, ${this.bullets.length} bullets`);
                this.lastFishCount = this.fishes.length;
                this.lastBulletCount = this.bullets.length;
            }
        }

        // 更新統計顯示
        document.getElementById('renderFishCount').textContent = this.fishes.length;
        document.getElementById('renderBulletCount').textContent = this.bullets.length;
    }

    /**
     * 開始渲染循環
     */
    start() {
        if (this.isRunning) return;

        this.isRunning = true;
        console.log('GameRenderer started');
        this.animate();
    }

    /**
     * 停止渲染循環
     */
    stop() {
        this.isRunning = false;
        if (this.animationId) {
            cancelAnimationFrame(this.animationId);
            this.animationId = null;
        }
        console.log('GameRenderer stopped');
    }

    /**
     * 清空畫面
     */
    clear() {
        this.fishes = [];
        this.bullets = [];
        this.formations = [];
        this.players.clear();
        this.currentPlayerId = null;
        this.ctx.clearRect(0, 0, this.width, this.height);
    }

    /**
     * 主動畫循環
     */
    animate() {
        if (!this.isRunning) return;

        // 清空畫布
        this.ctx.clearRect(0, 0, this.width, this.height);

        // 繪製遊戲對象
        this.drawFormations();
        this.drawFishes();
        this.drawBullets();
        this.drawCannons();

        // 調試：顯示當前有多少對象需要繪製
        if (this.fishes.length > 0 || this.bullets.length > 0 || this.players.size > 0) {
            this.ctx.save();
            this.ctx.fillStyle = 'rgba(255, 255, 255, 0.7)';
            this.ctx.font = '12px monospace';
            this.ctx.fillText(`Drawing: ${this.fishes.length} fish, ${this.bullets.length} bullets, ${this.players.size} players`, 10, this.height - 10);
            this.ctx.restore();
        }

        // 更新 FPS
        this.updateFPS();

        // 繼續動畫循環
        this.animationId = requestAnimationFrame(() => this.animate());
    }

    /**
     * 繪製魚群陣型輔助線和路徑可視化
     */
    drawFormations() {
        if (this.formations.length === 0) return;

        this.ctx.save();

        this.formations.forEach(formation => {
            // 阵型类型颜色映射
            const formationColors = {
                'v_formation': 'rgba(255, 100, 100, 0.4)',
                'line': 'rgba(100, 255, 100, 0.4)',
                'circle': 'rgba(100, 100, 255, 0.4)',
                'triangle': 'rgba(255, 255, 100, 0.4)',
                'diamond': 'rgba(255, 100, 255, 0.4)',
                'wave': 'rgba(100, 255, 255, 0.4)',
                'spiral': 'rgba(255, 150, 50, 0.4)'
            };
            const formationColor = formationColors[formation.type] || 'rgba(255, 255, 255, 0.3)';

            // 繪製陣型範圍
            this.ctx.beginPath();
            this.ctx.arc(formation.centerX, formation.centerY, 60, 0, Math.PI * 2);
            this.ctx.strokeStyle = formationColor;
            this.ctx.lineWidth = 2;
            this.ctx.setLineDash([5, 5]);
            this.ctx.stroke();

            // 繪製陣型中心點
            this.ctx.beginPath();
            this.ctx.arc(formation.centerX, formation.centerY, 6, 0, Math.PI * 2);
            this.ctx.fillStyle = formationColor.replace('0.4', '0.8');
            this.ctx.fill();
            this.ctx.strokeStyle = 'white';
            this.ctx.lineWidth = 2;
            this.ctx.setLineDash([]);
            this.ctx.stroke();

            // 繪製進度條
            const progressBarWidth = 80;
            const progressBarHeight = 6;
            const progressBarX = formation.centerX - progressBarWidth / 2;
            const progressBarY = formation.centerY - 75;

            // 進度條背景
            this.ctx.fillStyle = 'rgba(0, 0, 0, 0.5)';
            this.ctx.fillRect(progressBarX, progressBarY, progressBarWidth, progressBarHeight);

            // 進度條填充
            const progress = formation.progress || 0;
            this.ctx.fillStyle = formationColor.replace('0.4', '0.9');
            this.ctx.fillRect(progressBarX, progressBarY, progressBarWidth * progress, progressBarHeight);

            // 進度條邊框
            this.ctx.strokeStyle = 'white';
            this.ctx.lineWidth = 1;
            this.ctx.strokeRect(progressBarX, progressBarY, progressBarWidth, progressBarHeight);

            // 繪製陣型類型標籤
            this.ctx.font = 'bold 12px monospace';
            this.ctx.textAlign = 'center';
            this.ctx.fillStyle = 'white';
            this.ctx.strokeStyle = 'black';
            this.ctx.lineWidth = 3;
            const typeText = this.getFormationTypeName(formation.type);
            this.ctx.strokeText(typeText, formation.centerX, progressBarY - 5);
            this.ctx.fillText(typeText, formation.centerX, progressBarY - 5);

            // 繪製魚數量
            this.ctx.font = '10px monospace';
            const fishCountText = `${formation.fishIds.length} 魚`;
            this.ctx.strokeText(fishCountText, formation.centerX, formation.centerY + 80);
            this.ctx.fillText(fishCountText, formation.centerX, formation.centerY + 80);

            // 繪製連接線到陣型中的魚
            if (formation.fishIds && formation.fishIds.length > 0) {
                this.ctx.strokeStyle = formationColor.replace('0.4', '0.2');
                this.ctx.lineWidth = 1;
                this.ctx.setLineDash([2, 3]);

                formation.fishIds.forEach(fishId => {
                    const fish = this.fishes.find(f => f.id === fishId);
                    if (fish) {
                        this.ctx.beginPath();
                        this.ctx.moveTo(formation.centerX, formation.centerY);
                        this.ctx.lineTo(fish.x, fish.y);
                        this.ctx.stroke();
                    }
                });
            }
        });

        this.ctx.restore();
    }

    /**
     * 獲取陣型類型的中文名稱
     */
    getFormationTypeName(type) {
        const typeNames = {
            'v_formation': 'V字陣',
            'line': '直線陣',
            'circle': '圓形陣',
            'triangle': '三角陣',
            'diamond': '菱形陣',
            'wave': '波浪陣',
            'spiral': '螺旋陣'
        };
        return typeNames[type] || type;
    }

    /**
     * 繪製魚類
     */
    drawFishes() {
        if (this.fishes.length === 0) return;

        this.fishes.forEach(fish => {
            // 檢查魚是否在畫布範圍內（擴展範圍以顯示部分在外的魚）
            if (fish.x < -100 || fish.x > this.width + 100 ||
                fish.y < -100 || fish.y > this.height + 100) {
                return; // 跳過畫布外的魚
            }

            this.ctx.save();

            // 移動到魚的位置
            this.ctx.translate(fish.x, fish.y);
            this.ctx.rotate(fish.direction);

            // 獲取魚的顏色
            const color = this.fishColors[fish.type] || '#FFD700';

            // 繪製魚身 (簡單橢圓)
            const fishWidth = 20 + fish.type * 5; // 根據類型調整大小
            const fishHeight = 12 + fish.type * 3;

            // 魚身
            this.ctx.beginPath();
            this.ctx.ellipse(0, 0, fishWidth, fishHeight, 0, 0, Math.PI * 2);
            this.ctx.fillStyle = color;
            this.ctx.fill();
            this.ctx.strokeStyle = 'rgba(0, 0, 0, 0.3)';
            this.ctx.lineWidth = 1;
            this.ctx.stroke();

            // 魚尾
            this.ctx.beginPath();
            this.ctx.moveTo(-fishWidth, 0);
            this.ctx.lineTo(-fishWidth - 10, -8);
            this.ctx.lineTo(-fishWidth - 10, 8);
            this.ctx.closePath();
            this.ctx.fillStyle = color;
            this.ctx.fill();
            this.ctx.stroke();

            // 魚眼
            this.ctx.beginPath();
            this.ctx.arc(fishWidth * 0.6, -fishHeight * 0.3, 3, 0, Math.PI * 2);
            this.ctx.fillStyle = '#000000';
            this.ctx.fill();

            this.ctx.restore();

            // 繪製血量條
            if (fish.health < fish.maxHealth) {
                this.ctx.save();
                const barWidth = 30;
                const barHeight = 4;
                const barX = fish.x - barWidth / 2;
                const barY = fish.y - 25;

                // 背景
                this.ctx.fillStyle = 'rgba(0, 0, 0, 0.5)';
                this.ctx.fillRect(barX, barY, barWidth, barHeight);

                // 血量
                const healthPercent = fish.health / fish.maxHealth;
                this.ctx.fillStyle = healthPercent > 0.5 ? '#4CAF50' : '#F44336';
                this.ctx.fillRect(barX, barY, barWidth * healthPercent, barHeight);

                this.ctx.restore();
            }

            // 繪製魚的價值 (小號文字)
            this.ctx.save();
            this.ctx.font = '10px Arial';
            this.ctx.fillStyle = 'rgba(255, 255, 255, 0.8)';
            this.ctx.textAlign = 'center';
            this.ctx.fillText(`${fish.value}`, fish.x, fish.y + 30);
            this.ctx.restore();
        });
    }

    /**
     * 繪製子彈
     */
    drawBullets() {
        if (this.bullets.length === 0) return;

        this.bullets.forEach(bullet => {
            // 檢查子彈是否在畫布範圍內
            if (bullet.x < -50 || bullet.x > this.width + 50 ||
                bullet.y < -50 || bullet.y > this.height + 50) {
                return; // 跳過畫布外的子彈
            }

            this.ctx.save();

            // 移動到子彈位置
            this.ctx.translate(bullet.x, bullet.y);
            this.ctx.rotate(bullet.direction);

            // 繪製子彈 (火箭形狀)
            const bulletWidth = 8;
            const bulletHeight = 20;

            // 子彈頭部 (三角形)
            this.ctx.beginPath();
            this.ctx.moveTo(0, -bulletHeight / 2);
            this.ctx.lineTo(bulletWidth / 2, 0);
            this.ctx.lineTo(-bulletWidth / 2, 0);
            this.ctx.closePath();
            this.ctx.fillStyle = '#FF4444';
            this.ctx.fill();

            // 子彈身體
            this.ctx.fillStyle = '#FFAA00';
            this.ctx.fillRect(-bulletWidth / 2, 0, bulletWidth, bulletHeight / 2);

            // 子彈尾焰
            this.ctx.beginPath();
            this.ctx.moveTo(-bulletWidth / 2, bulletHeight / 2);
            this.ctx.lineTo(0, bulletHeight);
            this.ctx.lineTo(bulletWidth / 2, bulletHeight / 2);
            this.ctx.closePath();
            this.ctx.fillStyle = '#FF6600';
            this.ctx.fill();

            // 光暈效果
            const gradient = this.ctx.createRadialGradient(0, 0, 0, 0, 0, bulletHeight);
            gradient.addColorStop(0, 'rgba(255, 200, 0, 0.8)');
            gradient.addColorStop(1, 'rgba(255, 200, 0, 0)');
            this.ctx.beginPath();
            this.ctx.arc(0, 0, bulletHeight, 0, Math.PI * 2);
            this.ctx.fillStyle = gradient;
            this.ctx.fill();

            this.ctx.restore();
        });
    }

    /**
     * 更新 FPS 顯示
     */
    updateFPS() {
        this.frameCount++;
        const now = Date.now();
        const elapsed = now - this.lastFpsUpdate;

        if (elapsed >= 1000) {
            this.fps = Math.round((this.frameCount * 1000) / elapsed);
            this.frameCount = 0;
            this.lastFpsUpdate = now;
            document.getElementById('fpsDisplay').textContent = this.fps;
        }
    }

    /**
     * 測試函數 - 添加測試魚類和子彈
     */
    addTestData() {
        console.log('[Renderer] Adding test data...');

        // 添加測試魚類
        this.fishes = [
            { id: 'test-fish-1', type: 1, x: 200, y: 200, direction: 0, speed: 2, health: 100, maxHealth: 100, value: 10 },
            { id: 'test-fish-2', type: 2, x: 400, y: 300, direction: Math.PI / 4, speed: 3, health: 80, maxHealth: 100, value: 20 },
            { id: 'test-fish-3', type: 3, x: 600, y: 400, direction: Math.PI / 2, speed: 1, health: 100, maxHealth: 100, value: 30 },
            { id: 'test-fish-4', type: 4, x: 800, y: 250, direction: -Math.PI / 4, speed: 2.5, health: 50, maxHealth: 100, value: 40 },
        ];

        // 添加測試子彈
        this.bullets = [
            { id: 'test-bullet-1', playerId: 'player1', x: 300, y: 500, direction: -Math.PI / 2, speed: 5, power: 50 },
            { id: 'test-bullet-2', playerId: 'player2', x: 700, y: 100, direction: Math.PI / 2, speed: 6, power: 75 },
        ];

        // 添加4個測試玩家（砲台）到不同座位
        this.setCurrentPlayer('player1');
        this.addPlayer('player1', 0);  // 座位0 - 底部
        this.addPlayer('player2', 1);  // 座位1 - 頂部
        this.addPlayer('player3', 2);  // 座位2 - 左側
        this.addPlayer('player4', 3);  // 座位3 - 右側

        // 更新統計
        document.getElementById('renderFishCount').textContent = this.fishes.length;
        document.getElementById('renderBulletCount').textContent = this.bullets.length;

        console.log('[Renderer] Test data added:', this.fishes.length, 'fishes,', this.bullets.length, 'bullets', this.players.size, 'players');
        console.log('[Renderer] Players in different seats:');
        this.players.forEach((player, id) => {
            console.log(`  ${id}: seat ${player.seatId}, angle ${(player.angle * 180 / Math.PI).toFixed(1)}°`);
        });
    }

    /**
     * 設置當前玩家ID
     */
    setCurrentPlayer(playerId) {
        this.currentPlayerId = playerId;
        console.log('[Renderer] Current player set to:', playerId);
    }

    /**
     * 添加玩家（砲台）
     * @param {string} playerId - 玩家ID
     * @param {number} [seatId] - 座位ID (0-3)，如果不提供則自動分配
     */
    addPlayer(playerId, seatId) {
        // 🔧 檢查是否已存在
        if (this.players.has(playerId)) {
            const existingPlayer = this.players.get(playerId);
            console.warn(`[Renderer] Player ${playerId} already exists at seat ${existingPlayer.seatId}. Skipping add.`);
            return;
        }

        // 如果提供了座位ID，使用座位ID；否則使用當前玩家數量作為索引
        const index = seatId !== undefined ? seatId : this.players.size;
        const positionData = this.getCannonPosition(index);

        this.players.set(playerId, {
            id: playerId,
            position: { x: positionData.x, y: positionData.y },
            cannonType: 1,
            level: 1,
            angle: positionData.angle,  // 使用座位對應的初始角度
            seatId: index               // 保存座位ID
        });

        console.log(`[Renderer] ✓ Player added: ${playerId} at seat ${index}, position (${positionData.x.toFixed(1)}, ${positionData.y.toFixed(1)}), angle ${(positionData.angle * 180 / Math.PI).toFixed(1)}°`);
        console.log(`[Renderer] Total players: ${this.players.size}, Current player: ${this.currentPlayerId}`);
    }

    /**
     * 移除玩家
     */
    removePlayer(playerId) {
        if (this.players.has(playerId)) {
            const player = this.players.get(playerId);
            console.log(`[Renderer] Removing player ${playerId} from seat ${player.seatId}`);
            this.players.delete(playerId);

            // 🔧 不要重新分配其他玩家的位置！座位系統應該保持固定
            // this.reassignPlayerPositions();
        }
    }

    /**
     * 重新分配所有玩家位置
     */
    reassignPlayerPositions() {
        const playerIds = Array.from(this.players.keys());
        playerIds.forEach((playerId, index) => {
            const player = this.players.get(playerId);
            const positionData = this.getCannonPosition(index);
            player.position = { x: positionData.x, y: positionData.y };
            player.angle = positionData.angle;  // 更新角度
            player.seatId = index;               // 更新座位ID
        });
    }

    /**
     * 獲取砲台位置和初始方向（根據玩家索引或座位ID）
     * @param {number} playerIndex - 玩家索引或座位ID (0-3)
     * @returns {{x: number, y: number, angle: number}} 位置和初始角度
     */
    getCannonPosition(playerIndex) {
        // 捕魚遊戲典型佈局：
        // - 座位0（索引0）：底部中央 - 向上發射
        // - 座位1（索引1）：頂部中央 - 向下發射
        // - 座位2（索引2）：左側中央 - 向右發射
        // - 座位3（索引3）：右側中央 - 向左發射

        const centerX = this.width / 2;
        const centerY = this.height / 2;
        const margin = 50;

        const positions = [
            { x: centerX, y: this.height - margin, angle: -Math.PI / 2 },  // 底部 - 向上 (-90°)
            { x: centerX, y: margin, angle: Math.PI / 2 },                 // 頂部 - 向下 (90°)
            { x: margin, y: centerY, angle: 0 },                           // 左側 - 向右 (0°)
            { x: this.width - margin, y: centerY, angle: Math.PI }         // 右側 - 向左 (180°)
        ];

        return positions[playerIndex % positions.length];
    }

    /**
     * 更新玩家砲台角度（根據滑鼠位置）
     */
    updateCannonAngle(playerId, targetX, targetY) {
        const player = this.players.get(playerId);
        if (player) {
            const dx = targetX - player.position.x;
            const dy = targetY - player.position.y;
            const newAngle = Math.atan2(dy, dx);

            // 調試：每100幀記錄一次
            if (this.frameCount % 100 === 0) {
                console.log(`[Renderer] Updating angle for ${playerId}: (${targetX.toFixed(0)}, ${targetY.toFixed(0)}) -> ${(newAngle * 180 / Math.PI).toFixed(1)}°`);
            }

            player.angle = newAngle;
        } else {
            if (this.frameCount % 100 === 0) {
                console.warn(`[Renderer] Player ${playerId} not found in players map`);
            }
        }
    }

    /**
     * 更新玩家砲台類型
     */
    updateCannonType(playerId, cannonType, level) {
        const player = this.players.get(playerId);
        if (player) {
            player.cannonType = cannonType;
            player.level = level;
            console.log(`[Renderer] Cannon updated for ${playerId}: type=${cannonType}, level=${level}`);
        }
    }

    /**
     * 獲取砲口位置（用於子彈發射）
     */
    getBarrelEndPosition(playerId) {
        const player = this.players.get(playerId);
        if (!player) {
            return null;
        }

        const barrelLength = 40 + player.level * 5;
        return {
            x: player.position.x + Math.cos(player.angle) * barrelLength,
            y: player.position.y + Math.sin(player.angle) * barrelLength,
            angle: player.angle,
            barrelLength: barrelLength
        };
    }

    /**
     * 繪製所有砲台
     */
    drawCannons() {
        // 調試：每100幀記錄一次砲台數量
        if (this.frameCount % 100 === 0 && this.players.size > 0) {
            console.log(`[Renderer] Drawing ${this.players.size} cannons`);
            this.players.forEach((player, playerId) => {
                console.log(`  - ${playerId}: seat ${player.seatId}, pos (${player.position.x.toFixed(0)}, ${player.position.y.toFixed(0)}), angle ${(player.angle * 180 / Math.PI).toFixed(1)}°`);
            });
        }

        this.players.forEach((player, playerId) => {
            const isCurrentPlayer = playerId === this.currentPlayerId;
            this.drawCannon(player, isCurrentPlayer);
        });
    }

    /**
     * 繪製單個砲台
     */
    drawCannon(player, isCurrentPlayer) {
        this.ctx.save();

        const { x, y } = player.position;
        const angle = player.angle;

        // 移動到砲台位置
        this.ctx.translate(x, y);
        this.ctx.rotate(angle);

        // 砲台底座
        const baseRadius = 30;
        this.ctx.beginPath();
        this.ctx.arc(0, 0, baseRadius, 0, Math.PI * 2);
        this.ctx.fillStyle = isCurrentPlayer ? '#4CAF50' : '#607D8B';
        this.ctx.fill();
        this.ctx.strokeStyle = isCurrentPlayer ? '#2E7D32' : '#455A64';
        this.ctx.lineWidth = 3;
        this.ctx.stroke();

        // 砲管
        const barrelLength = 40 + player.level * 5;
        const barrelWidth = 12 + player.level * 2;

        this.ctx.fillStyle = isCurrentPlayer ? '#66BB6A' : '#78909C';
        this.ctx.fillRect(0, -barrelWidth / 2, barrelLength, barrelWidth);
        this.ctx.strokeStyle = isCurrentPlayer ? '#2E7D32' : '#455A64';
        this.ctx.lineWidth = 2;
        this.ctx.strokeRect(0, -barrelWidth / 2, barrelLength, barrelWidth);

        // 砲口
        this.ctx.beginPath();
        this.ctx.arc(barrelLength, 0, barrelWidth / 2 + 2, 0, Math.PI * 2);
        this.ctx.fillStyle = isCurrentPlayer ? '#43A047' : '#546E7A';
        this.ctx.fill();
        this.ctx.stroke();

        // 調試：繪製砲口中心點（紅色小圓點）
        if (isCurrentPlayer) {
            this.ctx.beginPath();
            this.ctx.arc(barrelLength, 0, 3, 0, Math.PI * 2);
            this.ctx.fillStyle = '#FF0000';
            this.ctx.fill();
        }

        this.ctx.restore();

        // 獲取座位位置標籤偏移
        const seatId = player.seatId !== undefined ? player.seatId : -1;
        let labelOffsetX = 0, labelOffsetY = -45;

        // 根據座位位置調整標籤位置
        if (seatId === 0) {
            // 底部座位 - 標籤在上方
            labelOffsetY = -45;
        } else if (seatId === 1) {
            // 頂部座位 - 標籤在下方
            labelOffsetY = 60;
        } else if (seatId === 2) {
            // 左側座位 - 標籤在右方
            labelOffsetX = 50;
            labelOffsetY = 0;
        } else if (seatId === 3) {
            // 右側座位 - 標籤在左方
            labelOffsetX = -50;
            labelOffsetY = 0;
        }

        // 繪製座位標籤
        this.ctx.save();
        this.ctx.font = 'bold 10px Arial';
        this.ctx.textAlign = 'center';
        this.ctx.fillStyle = 'rgba(255, 255, 255, 0.6)';
        this.ctx.strokeStyle = 'rgba(0, 0, 0, 0.8)';
        this.ctx.lineWidth = 2;
        const seatLabel = seatId >= 0 ? `🪑 座位 ${seatId + 1}` : '未分配';
        this.ctx.strokeText(seatLabel, x + labelOffsetX, y + labelOffsetY - 15);
        this.ctx.fillText(seatLabel, x + labelOffsetX, y + labelOffsetY - 15);
        this.ctx.restore();

        // 繪製玩家ID標籤
        this.ctx.save();
        this.ctx.font = 'bold 12px Arial';
        this.ctx.fillStyle = isCurrentPlayer ? '#4CAF50' : '#FFFFFF';
        this.ctx.textAlign = 'center';
        this.ctx.strokeStyle = '#000000';
        this.ctx.lineWidth = 3;
        this.ctx.strokeText(player.id, x + labelOffsetX, y + labelOffsetY);
        this.ctx.fillText(player.id, x + labelOffsetX, y + labelOffsetY);

        // 顯示等級
        if (player.level > 1) {
            this.ctx.font = '10px Arial';
            this.ctx.fillStyle = '#FFD700';
            this.ctx.strokeText(`Lv.${player.level}`, x + labelOffsetX, y + labelOffsetY + 15);
            this.ctx.fillText(`Lv.${player.level}`, x + labelOffsetX, y + labelOffsetY + 15);
        }
        this.ctx.restore();
    }
}

// 初始化渲染器
document.addEventListener('DOMContentLoaded', () => {
    try {
        // 創建全局遊戲渲染器實例
        window.gameRenderer = new GameRenderer('gameCanvas');
        console.log('Game renderer ready and attached to window');
        console.log('gameRenderer.isRunning:', window.gameRenderer.isRunning);

        // 添加滑鼠移動事件，讓砲台跟隨滑鼠
        const canvas = document.getElementById('gameCanvas');
        if (canvas) {
            let mouseMoveCount = 0; // 用於控制日誌頻率

            canvas.addEventListener('mousemove', (event) => {
                mouseMoveCount++;

                // 🔧 只有在玩家已加入渲染器時才更新角度（即已選擇座位）
                if (window.gameRenderer && gameRenderer.isRunning && gameRenderer.currentPlayerId) {
                    const player = gameRenderer.players.get(gameRenderer.currentPlayerId);
                    if (player) {
                        const rect = canvas.getBoundingClientRect();
                        const mouseX = event.clientX - rect.left;
                        const mouseY = event.clientY - rect.top;
                        gameRenderer.updateCannonAngle(gameRenderer.currentPlayerId, mouseX, mouseY);
                    } else {
                        // 玩家不存在，每200次記錄一次
                        if (mouseMoveCount % 200 === 0) {
                            const playersArray = Array.from(gameRenderer.players.keys());
                            console.warn(`[Renderer] ⚠️ currentPlayerId="${gameRenderer.currentPlayerId}" not found!`);
                            console.warn(`[Renderer] Players in renderer: [${playersArray.join(', ')}]`);
                            console.warn(`[Renderer] Checking exact match:`, playersArray.map(p => `"${p}" === "${gameRenderer.currentPlayerId}" ? ${p === gameRenderer.currentPlayerId}`));
                        }
                    }
                } else {
                    // 條件不滿足，每200次記錄一次
                    if (mouseMoveCount % 200 === 0) {
                        console.warn(`[Renderer] Mouse move but conditions not met: renderer=${!!window.gameRenderer}, running=${gameRenderer?.isRunning}, playerId="${gameRenderer?.currentPlayerId}"`);
                    }
                }
            });

            // 添加點擊事件：點擊發射子彈
            canvas.addEventListener('click', (event) => {
                if (window.gameRenderer && gameRenderer.isRunning && gameRenderer.currentPlayerId) {
                    const player = gameRenderer.players.get(gameRenderer.currentPlayerId);
                    if (player) {
                        const rect = canvas.getBoundingClientRect();
                        const clickX = event.clientX - rect.left;
                        const clickY = event.clientY - rect.top;
                        console.log(`[Renderer] Canvas clicked at (${clickX}, ${clickY}) - triggering fire`);

                        // 觸發開火按鈕點擊事件
                        const fireBulletBtn = document.getElementById('fireBulletBtn');
                        if (fireBulletBtn && !fireBulletBtn.disabled) {
                            fireBulletBtn.click();
                        } else {
                            console.log('[Renderer] Fire button disabled - please select a seat first');
                        }
                    } else {
                        console.log('[Renderer] Player not in renderer - please select a seat first');
                    }
                }
            });

            canvas.style.cursor = 'crosshair'; // 改變滑鼠指標樣式
        }
    } catch (error) {
        console.error('Failed to initialize game renderer:', error);
    }
});
