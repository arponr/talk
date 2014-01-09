var input, output, websocket;

function showMessage(m) {
    var p = document.createElement("p");
    p.innerHTML = m;
    MathJax.Hub.Queue(["Typeset", MathJax.Hub, p]);
    output.appendChild(p);
}

function onMessage(e) {
    showMessage(e.data);
}

function onClose() {
    showMessage("Connection closed.");
}

function sendMessage() {
    var m = input.value;
    input.value = "";
    websocket.send(m);
}

function onKey(e) {
    if (e.keyCode == 13) {
        sendMessage();
    }
}

function init() {
    input = document.getElementById("input");
    input.addEventListener("keyup", onKey, false);

    output = document.getElementById("output");

    var host = location.origin.replace(/^http/, 'ws')
    websocket = new WebSocket(host + '/socket');
    websocket.onmessage = onMessage;
    websocket.onclose = onClose;
}

window.addEventListener("load", init, false);
