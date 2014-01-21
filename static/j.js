var input, output, websocket;

function showMessage(m) {
    var msg = document.createElement("div");
    msg.className = "msg";
    var username = document.createElement("span");
    username.className = "username";
    username.innerHTML = m.Username + ":";
    msg.appendChild(username);
    var body = document.createElement("div");
    body.className = "body";
    body.innerHTML = m.Body;
    msg.appendChild(body);
    MathJax.Hub.Queue(["Typeset", MathJax.Hub, body]);
    var bottom = output.scrollTop == output.scrollHeight - output.offsetHeight;
    output.appendChild(msg);
    if (bottom) scroll();
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
        return false;
    }
}

function scroll() {
    output.scrollTop = output.scrollHeight;
}

window.onload = function() {
    input = document.getElementById("input");
    input.onkeydown = onKey;

    output = document.getElementById("msgs");
    MathJax.Hub.Register.StartupHook("End", scroll);
    MathJax.Hub.Config({
        "HTML-CSS": {
            scale: 95,
            availableFonts: [],
            webFont: "STIX-Web",
        }
    });

    var url = location.href.replace(/^http/, "ws").replace("thread", "socket")
    websocket = new WebSocket(url);
    websocket.onmessage = onMessage;

    var newthread = document.getElementById("newthread");
    newthread.style.display = "none";
    document.getElementById("plusicon").onclick = function() {
        if (newthread.style.display != "none") {
            newthread.style.display = "none";
        } else {
            newthread.style.display = "block";
            newthread.elements[0].focus();
        }
    };
};
