// Auth Client - 處理認證和錢包管理

// API 配置
const API_CONFIG = {
    adminBaseURL: 'http://localhost:6060',
    gameServerURL: 'ws://localhost:9090/ws',
};

// 應用狀態
const appState = {
    token: null,
    user: null,
    walletId: null,
    isGuest: false,
};

// ==================== 工具函數 ====================

// 顯示提示訊息
function showAlert(message, type = 'info') {
    const alertBox = document.getElementById('alertBox');
    if (!alertBox) return; // 元素不存在時直接返回

    alertBox.textContent = message;
    alertBox.className = `alert alert-${type} show`;

    setTimeout(() => {
        alertBox.classList.remove('show');
    }, 5000);
}

// 保存認證資訊到 localStorage
function saveAuth(token, user, isGuest = false) {
    localStorage.setItem('auth_token', token);
    localStorage.setItem('user_info', JSON.stringify(user));
    localStorage.setItem('is_guest', isGuest.toString());
    appState.token = token;
    appState.user = user;
    appState.isGuest = isGuest;
}

// 讀取認證資訊
function loadAuth() {
    const token = localStorage.getItem('auth_token');
    const userInfo = localStorage.getItem('user_info');
    const isGuest = localStorage.getItem('is_guest') === 'true';

    if (token && userInfo) {
        appState.token = token;
        appState.user = JSON.parse(userInfo);
        appState.isGuest = isGuest;
        return true;
    }
    return false;
}

// 清除認證資訊
function clearAuth() {
    localStorage.removeItem('auth_token');
    localStorage.removeItem('user_info');
    localStorage.removeItem('is_guest');
    localStorage.removeItem('wallet_id');
    appState.token = null;
    appState.user = null;
    appState.isGuest = false;
    appState.walletId = null;
}

// API 請求封裝
async function apiRequest(endpoint, options = {}) {
    const url = `${API_CONFIG.adminBaseURL}${endpoint}`;
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers,
    };

    if (appState.token) {
        headers['Authorization'] = `Bearer ${appState.token}`;
    }

    try {
        const response = await fetch(url, {
            ...options,
            headers,
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || data.message || '請求失敗');
        }

        return data;
    } catch (error) {
        console.error('API Request Error:', error);
        throw error;
    }
}

// ==================== 認證相關函數 ====================

// 註冊
async function register(username, password, nickname = '') {
    const data = await apiRequest('/api/v1/auth/register', {
        method: 'POST',
        body: JSON.stringify({
            username,
            password,
            nickname: nickname || username,
        }),
    });

    saveAuth(data.token, data.user, false);
    return data;
}

// 登入
async function login(username, password) {
    const data = await apiRequest('/api/v1/auth/login', {
        method: 'POST',
        body: JSON.stringify({
            username,
            password,
        }),
    });

    // 登入成功後需要獲取用戶資訊
    saveAuth(data.token, null, false);
    const userProfile = await getUserProfile();
    appState.user = userProfile.user;
    localStorage.setItem('user_info', JSON.stringify(userProfile.user));

    return data;
}

// 遊客登入
async function guestLogin() {
    const data = await apiRequest('/api/v1/auth/guest-login', {
        method: 'POST',
    });

    // 遊客登入後獲取用戶資訊
    saveAuth(data.token, null, true);
    const userProfile = await getUserProfile();
    appState.user = userProfile.user;
    localStorage.setItem('user_info', JSON.stringify(userProfile.user));

    return data;
}

// 獲取用戶資料
async function getUserProfile() {
    return await apiRequest('/api/v1/user/profile', {
        method: 'GET',
    });
}

// 登出
function logout() {
    clearAuth();
    showAuthPanel();
    showAlert('已登出', 'success');
}

// ==================== 錢包相關函數 ====================

// 獲取玩家的錢包列表
async function getPlayerWallets(userId) {
    return await apiRequest(`/admin/players/${userId}/wallets`, {
        method: 'GET',
    });
}

// 獲取錢包詳情
async function getWallet(walletId) {
    return await apiRequest(`/admin/wallets/${walletId}`, {
        method: 'GET',
    });
}

// 儲值
async function deposit(walletId, amount, description = '') {
    return await apiRequest(`/admin/wallets/${walletId}/deposit`, {
        method: 'POST',
        body: JSON.stringify({
            amount: parseFloat(amount),
            description: description || '用戶儲值',
            type: 'user_deposit',
        }),
    });
}

// 提款
async function withdraw(walletId, amount, description = '') {
    return await apiRequest(`/admin/wallets/${walletId}/withdraw`, {
        method: 'POST',
        body: JSON.stringify({
            amount: parseFloat(amount),
            description: description || '用戶提款',
            type: 'user_withdraw',
        }),
    });
}

// 獲取交易記錄
async function getTransactions(walletId, limit = 10, offset = 0) {
    return await apiRequest(`/admin/wallets/${walletId}/transactions?limit=${limit}&offset=${offset}`, {
        method: 'GET',
    });
}

// ==================== UI 控制函數 ====================

// 顯示認證面板
function showAuthPanel() {
    const authPanel = document.getElementById('authPanel');
    const mainPanel = document.getElementById('mainPanel');
    if (!authPanel || !mainPanel) return; // 不在 auth.html 頁面

    authPanel.style.display = 'flex';
    mainPanel.classList.remove('active');
}

// 顯示主面板
function showMainPanel() {
    const authPanel = document.getElementById('authPanel');
    const mainPanel = document.getElementById('mainPanel');
    if (!authPanel || !mainPanel) return; // 不在 auth.html 頁面

    authPanel.style.display = 'none';
    mainPanel.classList.add('active');
    updateUserInfo();
    loadWalletInfo();
}

// 更新用戶資訊顯示
function updateUserInfo() {
    if (!appState.user) return;

    const userName = document.getElementById('userName');
    const userAvatar = document.getElementById('userAvatar');
    const userStatus = document.getElementById('userStatus');

    // 如果元素不存在（不在 auth.html 頁面），直接返回
    if (!userName || !userAvatar || !userStatus) return;

    const displayName = appState.user.nickname || appState.user.username || '用戶';
    userName.textContent = displayName;
    userAvatar.textContent = displayName.charAt(0).toUpperCase();

    if (appState.isGuest) {
        userStatus.textContent = '遊客模式';
        userStatus.style.color = '#ff9800';
    } else {
        userStatus.textContent = '已登入';
        userStatus.style.color = '#4caf50';
    }
}

// 載入錢包資訊
async function loadWalletInfo() {
    try {
        if (!appState.user || !appState.user.id) {
            showAlert('無法獲取用戶資訊', 'error');
            return;
        }

        // 遊客模式不載入錢包（遊客沒有數據庫記錄）
        if (appState.isGuest) {
            console.log('遊客模式，跳過錢包載入');
            updateWalletBalance(0); // 遊客餘額顯示為 0
            return;
        }

        // 獲取玩家的錢包列表
        const walletsData = await getPlayerWallets(appState.user.id);

        if (walletsData.wallets && walletsData.wallets.length > 0) {
            // 使用第一個錢包（CNY）
            const wallet = walletsData.wallets.find(w => w.currency === 'CNY') || walletsData.wallets[0];
            appState.walletId = wallet.id;
            localStorage.setItem('wallet_id', wallet.id);

            // 更新餘額顯示
            updateWalletBalance(wallet.balance);
        } else {
            showAlert('尚未創建錢包，請聯繫管理員', 'info');
        }
    } catch (error) {
        console.error('載入錢包資訊失敗:', error);
        showAlert('載入錢包資訊失敗: ' + error.message, 'error');
    }
}

// 刷新錢包餘額
async function refreshWallet() {
    try {
        if (!appState.walletId) {
            showAlert('未找到錢包資訊', 'error');
            return;
        }

        const walletData = await getWallet(appState.walletId);
        updateWalletBalance(walletData.balance);
        showAlert('餘額已更新', 'success');
    } catch (error) {
        console.error('刷新錢包失敗:', error);
        showAlert('刷新失敗: ' + error.message, 'error');
    }
}

// 更新錢包餘額顯示
function updateWalletBalance(balance) {
    const balanceElement = document.getElementById('walletBalance');
    if (!balanceElement) return; // 元素不存在時直接返回

    balanceElement.textContent = parseFloat(balance).toFixed(2);
}

// 載入交易記錄
async function loadTransactions() {
    try {
        if (!appState.walletId) {
            showAlert('未找到錢包資訊', 'error');
            return;
        }

        const data = await getTransactions(appState.walletId, 20, 0);
        displayTransactions(data.transactions || []);
    } catch (error) {
        console.error('載入交易記錄失敗:', error);
        showAlert('載入交易記錄失敗: ' + error.message, 'error');
    }
}

// 顯示交易記錄
function displayTransactions(transactions) {
    const container = document.getElementById('transactionsList');

    if (transactions.length === 0) {
        container.innerHTML = '<p style="text-align: center; color: #999;">暫無交易記錄</p>';
        return;
    }

    container.innerHTML = transactions.map(tx => {
        const isPositive = tx.amount > 0;
        const typeClass = isPositive ? 'deposit' : 'withdraw';
        const typeText = isPositive ? '存入' : '提取';
        const amountText = isPositive ? `+${tx.amount}` : tx.amount;

        return `
            <div class="transaction-item">
                <div>
                    <span class="transaction-type ${typeClass}">${typeText}</span>
                    <p style="margin: 5px 0 0 0; color: #666; font-size: 0.9em;">${tx.description || '無備註'}</p>
                    <p style="margin: 5px 0 0 0; color: #999; font-size: 0.8em;">${new Date(tx.created_at).toLocaleString()}</p>
                </div>
                <div style="text-align: right;">
                    <p style="font-size: 1.2em; font-weight: bold; color: ${isPositive ? '#4caf50' : '#f44336'};">${amountText}</p>
                    <p style="color: #666; font-size: 0.9em;">餘額: ${tx.balance_after}</p>
                </div>
            </div>
        `;
    }).join('');
}

// ==================== Modal 控制 ====================

function openDepositModal() {
    document.getElementById('depositModal').classList.add('show');
}

function closeDepositModal() {
    document.getElementById('depositModal').classList.remove('show');
    document.getElementById('depositForm').reset();
}

function openWithdrawModal() {
    document.getElementById('withdrawModal').classList.add('show');
}

function closeWithdrawModal() {
    document.getElementById('withdrawModal').classList.remove('show');
    document.getElementById('withdrawForm').reset();
}

// 開始遊戲
function startGame() {
    // 跳轉到遊戲頁面
    window.location.href = 'index.html';
}

// ==================== 事件處理 ====================

document.addEventListener('DOMContentLoaded', () => {
    // 檢查是否在認證頁面（有 authPanel 元素）
    const isAuthPage = !!document.getElementById('authPanel');

    // 檢查是否已登入
    if (loadAuth() && isAuthPage) {
        showMainPanel();
    }

    // 以下事件監聽器只在認證頁面需要
    if (!isAuthPage) {
        return; // 不在認證頁面，跳過表單事件綁定
    }

    // Tab 切換
    document.querySelectorAll('.auth-tab').forEach(tab => {
        tab.addEventListener('click', () => {
            const tabName = tab.dataset.tab;

            // 更新 tab 樣式
            document.querySelectorAll('.auth-tab').forEach(t => t.classList.remove('active'));
            tab.classList.add('active');

            // 切換表單
            document.querySelectorAll('.auth-form').forEach(form => form.classList.remove('active'));
            if (tabName === 'login') {
                document.getElementById('loginForm').classList.add('active');
            } else if (tabName === 'register') {
                document.getElementById('registerForm').classList.add('active');
            } else if (tabName === 'guest') {
                document.getElementById('guestPanel').classList.add('active');
            }
        });
    });

    // 登入表單
    const loginForm = document.getElementById('loginForm');
    if (loginForm) {
        loginForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const username = document.getElementById('loginUsername').value;
        const password = document.getElementById('loginPassword').value;

        try {
            await login(username, password);
            showAlert('登入成功！', 'success');
            showMainPanel();
        } catch (error) {
            showAlert('登入失敗: ' + error.message, 'error');
        }
        });
    }

    // 註冊表單
    const registerForm = document.getElementById('registerForm');
    if (registerForm) {
        registerForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const username = document.getElementById('registerUsername').value;
        const password = document.getElementById('registerPassword').value;
        const passwordConfirm = document.getElementById('registerPasswordConfirm').value;
        const nickname = document.getElementById('registerNickname').value;

        // 驗證密碼
        if (password !== passwordConfirm) {
            showAlert('兩次輸入的密碼不一致', 'error');
            return;
        }

        try {
            await register(username, password, nickname);
            showAlert('註冊成功！', 'success');
            showMainPanel();
        } catch (error) {
            showAlert('註冊失敗: ' + error.message, 'error');
        }
        });
    }

    // 遊客登入
    const guestLoginBtn = document.getElementById('guestLoginBtn');
    if (guestLoginBtn) {
        guestLoginBtn.addEventListener('click', async () => {
        try {
            await guestLogin();
            showAlert('已進入遊客模式！', 'success');
            showMainPanel();
        } catch (error) {
            showAlert('遊客登入失敗: ' + error.message, 'error');
        }
        });
    }

    // 登出按鈕
    const logoutBtn = document.getElementById('logoutBtn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', logout);
    }

    // 儲值表單
    const depositForm = document.getElementById('depositForm');
    if (depositForm) {
        depositForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const amount = document.getElementById('depositAmount').value;
        const description = document.getElementById('depositDescription').value;

        try {
            await deposit(appState.walletId, amount, description);
            showAlert('儲值成功！', 'success');
            closeDepositModal();
            await refreshWallet();
            await loadTransactions();
        } catch (error) {
            showAlert('儲值失敗: ' + error.message, 'error');
        }
        });
    }

    // 提款表單
    const withdrawForm = document.getElementById('withdrawForm');
    if (withdrawForm) {
        withdrawForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const amount = document.getElementById('withdrawAmount').value;
        const description = document.getElementById('withdrawDescription').value;

        try {
            await withdraw(appState.walletId, amount, description);
            showAlert('提款成功！', 'success');
            closeWithdrawModal();
            await refreshWallet();
            await loadTransactions();
        } catch (error) {
            showAlert('提款失敗: ' + error.message, 'error');
        }
        });
    }

    // Modal 外部點擊關閉
    const depositModal = document.getElementById('depositModal');
    if (depositModal) {
        depositModal.addEventListener('click', (e) => {
            if (e.target.id === 'depositModal') {
                closeDepositModal();
            }
        });
    }

    const withdrawModal = document.getElementById('withdrawModal');
    if (withdrawModal) {
        withdrawModal.addEventListener('click', (e) => {
            if (e.target.id === 'withdrawModal') {
                closeWithdrawModal();
            }
        });
    }
});

// 導出給其他頁面使用
window.authClient = {
    getToken: () => appState.token,
    getUser: () => appState.user,
    isAuthenticated: () => !!appState.token,
    isGuest: () => appState.isGuest,
    logout,
    loadAuth,
};

// 自動載入認證資訊（用於其他頁面如 index.html）
if (typeof document !== 'undefined' && !document.getElementById('authPanel')) {
    // 不在 auth.html 頁面時，自動載入認證資訊
    loadAuth();
    console.log('[AuthClient] Auto-loaded auth state:', {
        authenticated: !!appState.token,
        user: appState.user?.nickname || appState.user?.username,
        isGuest: appState.isGuest
    });
}
