function prezero(x) {
    return x < 10 ? "0" + x : x
}

function printDate(t) {
    return t.getFullYear()
         + "-" + prezero(t.getMonth() + 1)
         + "-" + prezero(t.getDate());
}

function printTime(t) {
    var h = t.getHours();
    var m = t.getMinutes();
    var p;
    if (h > 12) {
        h -= 12;
        p = "pm";
    } else {
        p = "am";
    }
    return h + ":" + prezero(m) + p;
}

function fmtTime(time, fmt) {
    var t = new Date(time.getAttribute("datetime"));
    time.innerHTML = fmt.replace("d", printDate(t))
                        .replace("t", printTime(t));
}

function fmtTimes(el, fmt) {
    var times = el.getElementsByTagName("time");
    for (var i = 0; i < times.length; i++) {
        fmtTime(times[i], fmt);
    }
}

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

    fmtTimes(document.getElementById("right"), "d, t");

    var output = document.getElementById("msgs");
    MathJax.Hub.Register.StartupHook("End", function() { scroll(output); });

    var socket = websocket(location.href);
    socket.onmessage = function(e) {
        m = JSON.parse(e.data);
        var username = create("div", "username light", m.Username + ":", []);
        var time = create("time", "time light sans", "", []);
        time.setAttribute("datetime", m.Time);
        fmtTime(time, "d, t");
        var body = create("div", "body" + (m.Tex ? " math" : ""), m.Body, []);
        var msg = create("div", "msg", "", [username, time, body]);
        if (m.Tex) MathJax.Hub.Queue(["Typeset", MathJax.Hub, body]);
        var atBottom = output.scrollTop == bottom(output)
        output.appendChild(msg);
        if (atBottom) scroll(output);
    };

    var markdown = document.getElementById("markdown");
    var tex = document.getElementById("tex");
    var input = document.getElementById("input");
    input.onkeydown = function(e) {
        if (e.shiftKey && e.keyCode == 13) {
            var m = {
                "Body": input.value,
                "Markdown": markdown.checked,
                "Tex": tex.checked,
            };
            input.value = "";
            socket.send(JSON.stringify(m));
            e.preventDefault();
        }
    };
}

function rootLoad() {
    mainLoad();
}

function mainLoad() {
    MathJax.Hub.Config({
        tex2jax: {
            inlineMath: [['$','$']],
            processClass: "math",
            ignoreClass: "nomath",
        },
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

    var left = document.getElementById("left");

    fmtTimes(left, "d<br/>t");

    var items = left.getElementsByClassName("item");
    for (var i = 1; i < items.length; i++) {
        var socket = websocket(items[i].href);
        var lastmsg = items[i].getElementsByClassName("lastmsg")[0];
        var time = items[i].getElementsByClassName("time")[0];
        socket.onmessage = function(l, t) {
            return function(e) {
                var m = JSON.parse(e.data);
                var d = new Date(m.Time);
                l.innerHTML =  m.Username + ": ";
                l.appendChild(create("span", m.Tex ? "math" : "", m.Body, []))
                t.setAttribute("datetime", m.Time);
                fmtTime(t, "d<br/>t");
                if (m.Tex) MathJax.Hub.Queue(["Typeset", MathJax.Hub, l]);
            };
        }(lastmsg, time);
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
