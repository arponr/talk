function create(tag, cl, html, ch) {
    var el = document.createElement(tag)
    if (cl) el.className = cl;
    el.innerHTML = html
    for (var i = 0; i < ch.length; i++) {
        el.appendChild(ch[i]);
    }
    return el
}

function bottom(el) {
    return el.scrollHeight - el.offsetHeight;
}

function scroll(el) {
    el.scrollTop = bottom(el);
}

function websocket(url) {
    return new WebSocket(url.replace(/^http/, "ws").replace("thread", "socket"));
}

function threadLoad() {
    mainLoad();

    var output = document.getElementById("msgs");
    MathJax.Hub.Register.StartupHook("End", function() { scroll(output); });

    var socket = websocket(location.href);
    socket.onmessage = function(e) {
        m = JSON.parse(e.data);
        var username = create("div", "light", m.Username + ":", []);
        var body = create("div", "", m.Body, []);
        var msg = create("div", "msg", "", [username, body]);
        MathJax.Hub.Queue(["Typeset", MathJax.Hub, body]);
        var atBottom = output.scrollTop == bottom(output)
        output.appendChild(msg);
        if (atBottom) scroll(output);
    };

    var input = document.getElementById("input");
    input.onkeydown = function(e) {
        if (e.shiftKey && e.keyCode == 13) {
            var m = input.value;
            input.value = "";
            socket.send(m);
            e.preventDefault();
        }
    };
}

function rootLoad() {
    mainLoad();
}

function mainLoad() {
    MathJax.Hub.Config({
        tex2jax: {inlineMath: [['$','$']]},
        "HTML-CSS": {
            scale: 95,
            availableFonts: [],
            webFont: "STIX-Web",
        }
    });

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

    var threads = document.getElementById("left").getElementsByClassName("item");
    for (var i = 1; i < threads.length; i++) {
        var socket = websocket(threads[i].href);
        var lastmsg = threads[i].getElementsByClassName("lastmsg")[0];
        socket.onmessage = function(l) {
            return function(e) {
                m = JSON.parse(e.data);
                l.innerHTML =  m.Username + ": " + m.Body;
                MathJax.Hub.Queue(["Typeset", MathJax.Hub, l]);
            };
        }(lastmsg);
    }
}

function loginLoad() {
    var flag = true;
    var sw = document.getElementById("switch");
    var login = document.getElementById("login");
    var submit = document.getElementById("submit");
    var again = document.getElementById("again");
    sw.onclick = function() {
        if (flag) {
            sw.value = "already have an account?";
            login.action = "/register";
            submit.value = "register";
            again.style.display = "block";
            flag = false;
        } else {
            sw.value = "need to register?";
            login.action = "/login";
            submit.value = "login";
            again.style.display = "none";
            flag = true;
        }
    };
}

var load = {"loginpage": loginLoad, "rootpage": rootLoad, "threadpage": threadLoad};

window.onload = function() { load[document.body.id](); };
