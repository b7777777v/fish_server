/**
 * 優化的遊戲渲染器 - 使用插值和客戶端預測
 *
 * 優化內容：
 * 1. 線性插值（Lerp）平滑過渡對象位置
 * 2. 基於速度的客戶端預測
 * 3. 使用 Map 存儲對象減少 GC
 * 4. Delta time 計算
 * 5. 批量 DOM 更新
 */
class GameRendererOptimized {
    constructor(canvasId) {
        this.canvas = document.getElementById(canvasId);
        if (!this.canvas) {
            throw new Error(`Canvas element with id "${canvasId}" not found`);
        }

        this.ctx = this.canvas.getContext('2d');
        this.width = this.canvas.width;
        this.height = this.canvas.height;

        // 使用 Map 存儲遊戲對象（key = id）
        this.fishes = new Map();
        this.bullets = new Map();
        this.formations = new Map();
        this.players = new Map();
        this.currentPlayerId = null;

        // 插值設置
        this.interpol ationFactor = 0.3; // 插值強度 (0-1，越大越平滑但延遲越高)
        this.serverUpdateInterval = 1000 / 20; // 假設服務器 20 Hz

        // Delta time 追蹤
        this.lastFrameTime = performance.now();
        this.deltaTime = 0;

        // FPS 追蹤
        this.fps = 0;
        this.frameCount = 0;
        this.lastFpsUpdate = Date.now();

        // 動畫
        this.animationId = null;
        this.isRunning = false;

        // 魚的顏色映射
        this.fishColors = {
            1: '#FFD700', 2: '#FF6B6B', 3: '#4ECDC4', 4: '#95E1D3',
            5: '#F38181', 6: '#AA96DA', 7: '#FCBAD3', 8: '#FFFFD2',
        };

        // 批量 DOM 更新緩衝（減少重繪）
        this.domUpdateBuffer = {
            fishCount: 0,
            bulletCount: 0,
            needsUpdate: false
        };

        console.log('GameRendererOptimized initialized with interpolation');
    }

    /**
     * 更新遊戲狀態（從服務器）
     * 使用插值而不是直接替換
     */
    updateGameState(roomStateUpdate) {
        if (!roomStateUpdate) return;

        const now = performance.now();

        // 更新魚類 - 使用插值
        this.updateFishes(roomStateUpdate.getFishesList(), now);

        // 更新子彈 - 使用插值
        this.updateBullets(roomStateUpdate.getBulletsList(), now);

        // 更新陣型
        this.updateFormations(roomStateUpdate.getFormationsList());

        // 標記需要更新 DOM
        this.domUpdateBuffer.fishCount = this.fishes.size;
        this.domUpdateBuffer.bulletCount = this.bullets.size;
        this.domUpdateBuffer.needsUpdate = true;
    }

    /**
     * 更新魚類 - 使用插值平滑過渡
     */
    updateFishes(fishesList, timestamp) {
        const newFishIds = new Set();

        fishesList.forEach(fishData => {
            const fishId = fishData.getFishId();
            newFishIds.add(fishId);

            const position = fishData.getPosition();
            const targetX = position.getX();
            const targetY = position.getY();

            if (this.fishes.has(fishId)) {
                // 更新現有魚 - 設置目標位置進行插值
                const fish = this.fishes.get(fishId);
                fish.targetX = targetX;
                fish.targetY = targetY;
                fish.speed = fishData.getSpeed();
                fish.direction = fishData.getDirection();
                fish.health = fishData.getHealth();
                fish.maxHealth = fishData.getMaxHealth();
                fish.value = fishData.getValue();
                fish.lastServerUpdate = timestamp;
            } else {
                // 新魚 - 直接設置位置（無需插值）
                this.fishes.set(fishId, {
                    id: fishId,
                    type: fishData.getFishType(),
                    x: targetX,  // 當前渲染位置
                    y: targetY,
                    targetX: targetX,  // 目標位置
                    targetY: targetY,
                    direction: fishData.getDirection(),
                    speed: fishData.getSpeed(),
                    health: fishData.getHealth(),
                    maxHealth: fishData.getMaxHealth(),
                    value: fishData.getValue(),
                    lastServerUpdate: timestamp
                });
            }
        });

        // 移除不存在的魚
        for (const fishId of this.fishes.keys()) {
            if (!newFishIds.has(fishId)) {
                this.fishes.delete(fishId);
            }
        }
    }

    /**
     * 更新子彈 - 使用插值
     */
    updateBullets(bulletsList, timestamp) {
        const newBulletIds = new Set();

        bulletsList.forEach(bulletData => {
            const bulletId = bulletData.getBulletId();
            newBulletIds.add(bulletId);

            const position = bulletData.getPosition();
            const targetX = position.getX();
            const targetY = position.getY();

            if (this.bullets.has(bulletId)) {
                // 更新現有子彈
                const bullet = this.bullets.get(bulletId);
                bullet.targetX = targetX;
                bullet.targetY = targetY;
                bullet.speed = bulletData.getSpeed();
                bullet.direction = bulletData.getDirection();
                bullet.power = bulletData.getPower();
                bullet.lastServerUpdate = timestamp;
            } else {
                // 新子彈
                this.bullets.set(bulletId, {
                    id: bulletId,
                    playerId: bulletData.getPlayerId(),
                    x: targetX,
                    y: targetY,
                    targetX: targetX,
                    targetY: targetY,
                    direction: bulletData.getDirection(),
                    speed: bulletData.getSpeed(),
                    power: bulletData.getPower(),
                    lastServerUpdate: timestamp
                });
            }
        });

        // 移除不存在的子彈
        for (const bulletId of this.bullets.keys()) {
            if (!newBulletIds.has(bulletId)) {
                this.bullets.delete(bulletId);
            }
        }
    }

    /**
     * 更新陣型（不需要插值）
     */
    updateFormations(formationsList) {
        this.formations.clear();
        formationsList.forEach(formation => {
            this.formations.set(formation.getFormationId(), {
                id: formation.getFormationId(),
                type: formation.getFormationType(),
                centerX: formation.getCenterPosition().getX(),
                centerY: formation.getCenterPosition().getY(),
                progress: formation.getProgress(),
                fishIds: formation.getFishIdsList()
            });
        });
    }

    /**
     * 主動畫循環 - 使用插值和 delta time
     */
    animate(timestamp = performance.now()) {
        if (!this.isRunning) return;

        // 計算 delta time (秒)
        this.deltaTime = (timestamp - this.lastFrameTime) / 1000;
        this.lastFrameTime = timestamp;

        // 限制 delta time 防止大幅跳躍
        if (this.deltaTime > 0.1) this.deltaTime = 0.1;

        // 更新對象位置（插值）
        this.interpolateObjects();

        // 清空畫布
        this.ctx.clearRect(0, 0, this.width, this.height);

        // 繪製遊戲對象
        this.drawFormations();
        this.drawFishes();
        this.drawBullets();
        this.drawCannons();

        // 更新 FPS
        this.updateFPS();

        // 批量更新 DOM（每 10 幀一次）
        if (this.frameCount % 10 === 0 && this.domUpdateBuffer.needsUpdate) {
            document.getElementById('renderFishCount').textContent = this.domUpdateBuffer.fishCount;
            document.getElementById('renderBulletCount').textContent = this.domUpdateBuffer.bulletCount;
            this.domUpdateBuffer.needsUpdate = false;
        }

        // 繼續動畫循環
        this.animationId = requestAnimationFrame((ts) => this.animate(ts));
    }

    /**
     * 插值所有對象的位置
     * 使用線性插值（Lerp）平滑過渡
     */
    interpolateObjects() {
        const now = performance.now();

        // 插值魚類位置
        this.fishes.forEach(fish => {
            // 計算自上次服務器更新以來的時間
            const timeSinceUpdate = now - (fish.lastServerUpdate || now);

            // 如果服務器更新很久沒來，使用預測而不是插值
            if (timeSinceUpdate > this.serverUpdateInterval * 2) {
                // 外推：基於速度預測位置
                const predictDistance = fish.speed * this.deltaTime;
                fish.x += Math.cos(fish.direction) * predictDistance;
                fish.y += Math.sin(fish.direction) * predictDistance;
            } else {
                // 插值：平滑過渡到目標位置
                const lerpFactor = Math.min(1, this.interpolationFactor);
                fish.x += (fish.targetX - fish.x) * lerpFactor;
                fish.y += (fish.targetY - fish.y) * lerpFactor;
            }
        });

        // 插值子彈位置（子彈移動快，使用更激進的預測）
        this.bullets.forEach(bullet => {
            const timeSinceUpdate = now - (bullet.lastServerUpdate || now);

            if (timeSinceUpdate > this.serverUpdateInterval) {
                // 子彈使用外推
                const predictDistance = bullet.speed * this.deltaTime;
                bullet.x += Math.cos(bullet.direction) * predictDistance;
                bullet.y += Math.sin(bullet.direction) * predictDistance;
            } else {
                // 快速插值
                const lerpFactor = 0.5; // 子彈使用更快的插值
                bullet.x += (bullet.targetX - bullet.x) * lerpFactor;
                bullet.y += (bullet.targetY - bullet.y) * lerpFactor;
            }
        });
    }

    /**
     * 繪製魚類（優化版）
     */
    drawFishes() {
        if (this.fishes.size === 0) return;

        this.fishes.forEach(fish => {
            // 剔除不可見的魚
            if (fish.x < -100 || fish.x > this.width + 100 ||
                fish.y < -100 || fish.y > this.height + 100) {
                return;
            }

            this.ctx.save();
            this.ctx.translate(fish.x, fish.y);
            this.ctx.rotate(fish.direction);

            const color = this.fishColors[fish.type] || '#FFD700';
            const fishWidth = 20 + fish.type * 5;
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

            // 血量條
            if (fish.health < fish.maxHealth) {
                const barWidth = 30;
                const barHeight = 4;
                const barX = fish.x - barWidth / 2;
                const barY = fish.y - 25;

                this.ctx.fillStyle = 'rgba(0, 0, 0, 0.5)';
                this.ctx.fillRect(barX, barY, barWidth, barHeight);

                const healthPercent = fish.health / fish.maxHealth;
                this.ctx.fillStyle = healthPercent > 0.5 ? '#4CAF50' : '#F44336';
                this.ctx.fillRect(barX, barY, barWidth * healthPercent, barHeight);
            }

            // 價值標籤
            this.ctx.font = '10px Arial';
            this.ctx.fillStyle = 'rgba(255, 255, 255, 0.8)';
            this.ctx.textAlign = 'center';
            this.ctx.fillText(`${fish.value}`, fish.x, fish.y + 30);
        });
    }

    /**
     * 繪製子彈（優化版）
     */
    drawBullets() {
        if (this.bullets.size === 0) return;

        this.bullets.forEach(bullet => {
            // 剔除不可見的子彈
            if (bullet.x < -50 || bullet.x > this.width + 50 ||
                bullet.y < -50 || bullet.y > this.height + 50) {
                return;
            }

            this.ctx.save();
            this.ctx.translate(bullet.x, bullet.y);
            this.ctx.rotate(bullet.direction);

            const bulletWidth = 8;
            const bulletHeight = 20;

            // 子彈頭部
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

            this.ctx.restore();
        });
    }

    /**
     * 繪製陣型（複製原有邏輯）
     */
    drawFormations() {
        if (this.formations.size === 0) return;

        this.ctx.save();

        this.formations.forEach(formation => {
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

            // 陣型範圍
            this.ctx.beginPath();
            this.ctx.arc(formation.centerX, formation.centerY, 60, 0, Math.PI * 2);
            this.ctx.strokeStyle = formationColor;
            this.ctx.lineWidth = 2;
            this.ctx.setLineDash([5, 5]);
            this.ctx.stroke();

            // 陣型中心
            this.ctx.beginPath();
            this.ctx.arc(formation.centerX, formation.centerY, 6, 0, Math.PI * 2);
            this.ctx.fillStyle = formationColor.replace('0.4', '0.8');
            this.ctx.fill();
            this.ctx.strokeStyle = 'white';
            this.ctx.lineWidth = 2;
            this.ctx.setLineDash([]);
            this.ctx.stroke();

            // 進度條
            const progressBarWidth = 80;
            const progressBarHeight = 6;
            const progressBarX = formation.centerX - progressBarWidth / 2;
            const progressBarY = formation.centerY - 75;

            this.ctx.fillStyle = 'rgba(0, 0, 0, 0.5)';
            this.ctx.fillRect(progressBarX, progressBarY, progressBarWidth, progressBarHeight);

            const progress = formation.progress || 0;
            this.ctx.fillStyle = formationColor.replace('0.4', '0.9');
            this.ctx.fillRect(progressBarX, progressBarY, progressBarWidth * progress, progressBarHeight);

            this.ctx.strokeStyle = 'white';
            this.ctx.lineWidth = 1;
            this.ctx.strokeRect(progressBarX, progressBarY, progressBarWidth, progressBarHeight);

            // 陣型標籤
            this.ctx.font = 'bold 12px monospace';
            this.ctx.textAlign = 'center';
            this.ctx.fillStyle = 'white';
            this.ctx.strokeStyle = 'black';
            this.ctx.lineWidth = 3;
            const typeText = this.getFormationTypeName(formation.type);
            this.ctx.strokeText(typeText, formation.centerX, progressBarY - 5);
            this.ctx.fillText(typeText, formation.centerX, progressBarY - 5);
        });

        this.ctx.restore();
    }

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
     * 繪製砲台（複製原有邏輯）
     */
    drawCannons() {
        this.players.forEach((player, playerId) => {
            const isCurrentPlayer = playerId === this.currentPlayerId;
            this.drawCannon(player, isCurrentPlayer);
        });
    }

    drawCannon(player, isCurrentPlayer) {
        this.ctx.save();

        const { x, y } = player.position;
        const angle = player.angle;

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

        this.ctx.restore();

        // 玩家標籤
        const seatId = player.seatId !== undefined ? player.seatId : -1;
        let labelOffsetX = 0, labelOffsetY = -45;

        if (seatId === 0) labelOffsetY = -45;
        else if (seatId === 1) labelOffsetY = 60;
        else if (seatId === 2) { labelOffsetX = 50; labelOffsetY = 0; }
        else if (seatId === 3) { labelOffsetX = -50; labelOffsetY = 0; }

        this.ctx.font = 'bold 12px Arial';
        this.ctx.fillStyle = isCurrentPlayer ? '#4CAF50' : '#FFFFFF';
        this.ctx.textAlign = 'center';
        this.ctx.strokeStyle = '#000000';
        this.ctx.lineWidth = 3;
        this.ctx.strokeText(player.id, x + labelOffsetX, y + labelOffsetY);
        this.ctx.fillText(player.id, x + labelOffsetX, y + labelOffsetY);
    }

    // ========== 保留原有的功能方法 ==========

    start() {
        if (this.isRunning) return;
        this.isRunning = true;
        this.lastFrameTime = performance.now();
        console.log('GameRendererOptimized started with interpolation');
        this.animate();
    }

    stop() {
        this.isRunning = false;
        if (this.animationId) {
            cancelAnimationFrame(this.animationId);
            this.animationId = null;
        }
        console.log('GameRendererOptimized stopped');
    }

    clear() {
        this.fishes.clear();
        this.bullets.clear();
        this.formations.clear();
        this.players.clear();
        this.currentPlayerId = null;
        this.ctx.clearRect(0, 0, this.width, this.height);
    }

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

    setCurrentPlayer(playerId) {
        this.currentPlayerId = playerId;
        console.log('[RendererOptimized] Current player set to:', playerId);
    }

    addPlayer(playerId, seatId) {
        if (this.players.has(playerId)) {
            console.warn(`[RendererOptimized] Player ${playerId} already exists`);
            return;
        }

        const index = seatId !== undefined ? seatId : this.players.size;
        const positionData = this.getCannonPosition(index);

        this.players.set(playerId, {
            id: playerId,
            position: { x: positionData.x, y: positionData.y },
            cannonType: 1,
            level: 1,
            angle: positionData.angle,
            seatId: index
        });

        console.log(`[RendererOptimized] Player added: ${playerId} at seat ${index}`);
    }

    removePlayer(playerId) {
        this.players.delete(playerId);
    }

    getCannonPosition(playerIndex) {
        const centerX = this.width / 2;
        const centerY = this.height / 2;
        const margin = 50;

        const positions = [
            { x: centerX, y: this.height - margin, angle: -Math.PI / 2 },
            { x: centerX, y: margin, angle: Math.PI / 2 },
            { x: margin, y: centerY, angle: 0 },
            { x: this.width - margin, y: centerY, angle: Math.PI }
        ];

        return positions[playerIndex % positions.length];
    }

    updateCannonAngle(playerId, targetX, targetY) {
        const player = this.players.get(playerId);
        if (player) {
            const dx = targetX - player.position.x;
            const dy = targetY - player.position.y;
            player.angle = Math.atan2(dy, dx);
        }
    }

    updateCannonType(playerId, cannonType, level) {
        const player = this.players.get(playerId);
        if (player) {
            player.cannonType = cannonType;
            player.level = level;
        }
    }

    getBarrelEndPosition(playerId) {
        const player = this.players.get(playerId);
        if (!player) return null;

        const barrelLength = 40 + player.level * 5;
        return {
            x: player.position.x + Math.cos(player.angle) * barrelLength,
            y: player.position.y + Math.sin(player.angle) * barrelLength,
            angle: player.angle,
            barrelLength: barrelLength
        };
    }
}

// 初始化優化渲染器
document.addEventListener('DOMContentLoaded', () => {
    try {
        window.gameRenderer = new GameRendererOptimized('gameCanvas');
        console.log('✨ Optimized game renderer ready with interpolation!');

        // 添加滑鼠事件
        const canvas = document.getElementById('gameCanvas');
        if (canvas) {
            canvas.addEventListener('mousemove', (event) => {
                if (window.gameRenderer && gameRenderer.isRunning && gameRenderer.currentPlayerId) {
                    const player = gameRenderer.players.get(gameRenderer.currentPlayerId);
                    if (player) {
                        const rect = canvas.getBoundingClientRect();
                        const mouseX = event.clientX - rect.left;
                        const mouseY = event.clientY - rect.top;
                        gameRenderer.updateCannonAngle(gameRenderer.currentPlayerId, mouseX, mouseY);
                    }
                }
            });

            canvas.addEventListener('click', (event) => {
                if (window.gameRenderer && gameRenderer.isRunning && gameRenderer.currentPlayerId) {
                    const fireBulletBtn = document.getElementById('fireBulletBtn');
                    if (fireBulletBtn && !fireBulletBtn.disabled) {
                        fireBulletBtn.click();
                    }
                }
            });

            canvas.style.cursor = 'crosshair';
        }
    } catch (error) {
        console.error('Failed to initialize optimized renderer:', error);
    }
});
