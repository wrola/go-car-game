const canvas = document.getElementById('gameCanvas');
const ctx = canvas.getContext('2d');

let ws;
let gameState = null;
let keys = {
    w: false,
    a: false,
    s: false,
    d: false,
    up: false,
    down: false,
    left: false,
    right: false
};

const cameraOffset = { x: 0, y: 0 };

function connect() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;

    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
        console.log('Connected to server');
        document.getElementById('status').textContent = 'Connected - Race Started!';
        document.getElementById('status').className = 'connected';
    };

    ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        handleMessage(data);
    };

    ws.onclose = () => {
        console.log('Disconnected from server');
        document.getElementById('status').textContent = 'Disconnected';
        document.getElementById('status').className = 'disconnected';
        setTimeout(connect, 3000);
    };

    ws.onerror = (error) => {
        console.error('WebSocket error:', error);
    };
}

function handleMessage(data) {
    switch (data.type) {
        case 'connected':
            console.log('Local game initialized');
            break;

        case 'gameState':
            gameState = data;

            if (data.winner) {
                showWinner(data.winner);
            }

            updateCamera();
            render();
            break;
    }
}

function updateCamera() {
    if (!gameState || !gameState.players || gameState.players.length === 0) return;

    let avgX = 0;
    let avgY = 0;
    gameState.players.forEach(p => {
        avgX += p.x;
        avgY += p.y;
    });
    avgX /= gameState.players.length;
    avgY /= gameState.players.length;

    const targetX = canvas.width / 2 - avgX;
    const targetY = canvas.height / 2 - avgY;

    cameraOffset.x += (targetX - cameraOffset.x) * 0.1;
    cameraOffset.y += (targetY - cameraOffset.y) * 0.1;
}

function sendInput() {
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
            type: 'input',
            input: keys
        }));
    }
}

document.addEventListener('keydown', (e) => {
    let changed = false;

    switch (e.key.toLowerCase()) {
        case 'w':
            if (!keys.w) { keys.w = true; changed = true; }
            break;
        case 's':
            if (!keys.s) { keys.s = true; changed = true; }
            break;
        case 'a':
            if (!keys.a) { keys.a = true; changed = true; }
            break;
        case 'd':
            if (!keys.d) { keys.d = true; changed = true; }
            break;
        case 'arrowup':
            e.preventDefault();
            if (!keys.up) { keys.up = true; changed = true; }
            break;
        case 'arrowdown':
            e.preventDefault();
            if (!keys.down) { keys.down = true; changed = true; }
            break;
        case 'arrowleft':
            e.preventDefault();
            if (!keys.left) { keys.left = true; changed = true; }
            break;
        case 'arrowright':
            e.preventDefault();
            if (!keys.right) { keys.right = true; changed = true; }
            break;
    }

    if (changed) sendInput();
});

document.addEventListener('keyup', (e) => {
    let changed = false;

    switch (e.key.toLowerCase()) {
        case 'w':
            if (keys.w) { keys.w = false; changed = true; }
            break;
        case 's':
            if (keys.s) { keys.s = false; changed = true; }
            break;
        case 'a':
            if (keys.a) { keys.a = false; changed = true; }
            break;
        case 'd':
            if (keys.d) { keys.d = false; changed = true; }
            break;
        case 'arrowup':
            e.preventDefault();
            if (keys.up) { keys.up = false; changed = true; }
            break;
        case 'arrowdown':
            e.preventDefault();
            if (keys.down) { keys.down = false; changed = true; }
            break;
        case 'arrowleft':
            e.preventDefault();
            if (keys.left) { keys.left = false; changed = true; }
            break;
        case 'arrowright':
            e.preventDefault();
            if (keys.right) { keys.right = false; changed = true; }
            break;
    }

    if (changed) sendInput();
});

function render() {
    ctx.fillStyle = '#228B22';
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    if (!gameState) return;

    ctx.save();
    ctx.translate(cameraOffset.x, cameraOffset.y);

    drawTrack();

    if (gameState.checkpoints) {
        drawCheckpoints();
    }

    if (gameState.players) {
        gameState.players.forEach((player, index) => {
            drawPlayer(player, index);
        });
    }

    ctx.restore();

    drawUI();
}

function drawTrack() {
    if (!gameState.checkpoints || gameState.checkpoints.length === 0) return;

    ctx.strokeStyle = '#555';
    ctx.lineWidth = 120;
    ctx.lineCap = 'round';
    ctx.lineJoin = 'round';

    ctx.beginPath();
    ctx.moveTo(50, 300);

    gameState.checkpoints.forEach(cp => {
        ctx.lineTo(cp.x, cp.y);
    });
    ctx.stroke();

    ctx.strokeStyle = '#FFD700';
    ctx.lineWidth = 3;
    ctx.setLineDash([20, 20]);
    ctx.beginPath();
    ctx.moveTo(50, 300);
    gameState.checkpoints.forEach(cp => {
        ctx.lineTo(cp.x, cp.y);
    });
    ctx.stroke();
    ctx.setLineDash([]);
}

function drawCheckpoints() {
    gameState.checkpoints.forEach((cp, index) => {
        ctx.strokeStyle = index === gameState.checkpoints.length - 1 ? '#FFD700' : '#00FF00';
        ctx.lineWidth = 3;
        ctx.beginPath();
        ctx.arc(cp.x, cp.y, cp.radius, 0, Math.PI * 2);
        ctx.stroke();

        ctx.fillStyle = '#FFF';
        ctx.font = 'bold 20px Arial';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';

        if (index === gameState.checkpoints.length - 1) {
            ctx.fillText('FINISH', cp.x, cp.y);
        } else {
            ctx.fillText((index + 1).toString(), cp.x, cp.y);
        }
    });
}

function drawPlayer(player, index) {
    const colors = ['#FF6B6B', '#4ECDC4'];
    const x = player.x;
    const y = player.y;
    const angle = player.angle * Math.PI / 180;

    ctx.save();
    ctx.translate(x, y);
    ctx.rotate(angle);

    ctx.fillStyle = colors[index];
    ctx.fillRect(-20, -12, 40, 24);

    ctx.strokeStyle = '#000';
    ctx.lineWidth = 2;
    ctx.strokeRect(-20, -12, 40, 24);

    ctx.fillStyle = 'rgba(255, 255, 255, 0.5)';
    ctx.fillRect(5, -8, 10, 16);

    ctx.fillStyle = '#000';
    ctx.fillRect(-15, -12, 8, 4);
    ctx.fillRect(-15, 8, 8, 4);
    ctx.fillRect(7, -12, 8, 4);
    ctx.fillRect(7, 8, 8, 4);

    ctx.restore();

    ctx.fillStyle = '#000';
    ctx.font = 'bold 16px Arial';
    ctx.textAlign = 'center';
    ctx.fillText(`P${player.id} (${player.checkpoint}/${gameState.checkpoints.length})`, x, y - 30);
}

function drawUI() {
    if (!gameState || !gameState.players) return;

    const barWidth = 200;
    const barHeight = 30;
    const padding = 20;

    gameState.players.forEach((player, index) => {
        const colors = ['#FF6B6B', '#4ECDC4'];
        const y = padding + index * (barHeight + 10);

        ctx.fillStyle = '#333';
        ctx.fillRect(padding, y, barWidth, barHeight);

        const progress = player.checkpoint / gameState.checkpoints.length;
        ctx.fillStyle = colors[index];
        ctx.fillRect(padding, y, barWidth * progress, barHeight);

        ctx.strokeStyle = '#FFF';
        ctx.lineWidth = 2;
        ctx.strokeRect(padding, y, barWidth, barHeight);

        ctx.fillStyle = '#FFF';
        ctx.font = 'bold 14px Arial';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        ctx.fillText(`Player ${player.id}: ${player.checkpoint}/${gameState.checkpoints.length}`,
                     padding + barWidth / 2, y + barHeight / 2);
    });
}

function showWinner(winnerId) {
    const winnerDiv = document.getElementById('winner');
    winnerDiv.innerHTML = `<div class="gold">Player ${winnerId} Wins!</div><div style="font-size: 0.6em; margin-top: 20px;">Congratulations!</div>`;
    winnerDiv.classList.remove('hidden');

    setTimeout(() => {
        location.reload();
    }, 5000);
}

connect();
render();
