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
        this.players = [];
        this.formations = [];

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
        if (!roomStateUpdate) return;

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
        this.bullets = roomStateUpdate.getBulletsList().map(bullet => ({
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

        // 更新 FPS
        this.updateFPS();

        // 繼續動畫循環
        this.animationId = requestAnimationFrame(() => this.animate());
    }

    /**
     * 繪製魚群陣型輔助線 (可選)
     */
    drawFormations() {
        this.ctx.save();
        this.ctx.strokeStyle = 'rgba(255, 255, 255, 0.2)';
        this.ctx.lineWidth = 1;
        this.ctx.setLineDash([5, 5]);

        this.formations.forEach(formation => {
            // 繪製陣型中心點
            this.ctx.beginPath();
            this.ctx.arc(formation.centerX, formation.centerY, 5, 0, Math.PI * 2);
            this.ctx.fillStyle = 'rgba(255, 255, 255, 0.5)';
            this.ctx.fill();

            // 繪製陣型範圍圓
            this.ctx.beginPath();
            this.ctx.arc(formation.centerX, formation.centerY, 50, 0, Math.PI * 2);
            this.ctx.stroke();
        });

        this.ctx.restore();
    }

    /**
     * 繪製魚類
     */
    drawFishes() {
        this.fishes.forEach(fish => {
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
        this.bullets.forEach(bullet => {
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
}

// 全局遊戲渲染器實例
let gameRenderer = null;

// 初始化渲染器
document.addEventListener('DOMContentLoaded', () => {
    try {
        gameRenderer = new GameRenderer('gameCanvas');
        console.log('Game renderer ready');
    } catch (error) {
        console.error('Failed to initialize game renderer:', error);
    }
});
