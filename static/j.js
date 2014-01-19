var input, output, websocket;

function showMessage(m) {
    var msg = document.createElement("div");
    msg.className = "msg";
    var username = document.createElement("div");
    username.className = "username";
    username.innerHTML = m.Username + ":";
    msg.appendChild(username);
    var body = document.createElement("div");
    body.className = "body";
    body.innerHTML = m.Body;
    msg.appendChild(body);
    MathJax.Hub.Queue(["Typeset", MathJax.Hub, body]);
    output.appendChild(msg);
    scroll();
}

function onMessage(e) {
    showMessage(JSON.parse(e.data));
}

function sendMessage() {
    var m = input.value;
    input.value = "";
    websocket.send(m);
}

function onKey(e) {
    if (e.shiftKey && e.keyCode == 13) {
        sendMessage();
    }
}

function scroll() {
    output.scrollTop = output.scrollHeight;
}

function init() {
    input = document.getElementById("input");
    input.onkeyup = onKey;

    output = document.getElementById("msgs");
    MathJax.Hub.Register.StartupHook("End", scroll);

    var url = location.href.replace(/^http/, "ws").replace("thread", "socket")
    websocket = new WebSocket(url);
    websocket.onmessage = onMessage;
}

window.onload = init
