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
    var t = new Date(time.attr("datetime"));
    time.html(fmt.replace("d", printDate(t))
                 .replace("t", printTime(t)));
}

function fmtTimes(el, fmt) {
    $(el).find("time").each(function() {
        fmtTime($(this), fmt);
    });
}

function create(tag, cl, html, ch) {
    var el = $(document.createElement(tag));
    if (cl) el.attr("class", cl);
    el.html(html);
    el.append(ch);
    return el
}

function bottom(el) {
    return el[0].scrollHeight - el[0].offsetHeight;
}

function scroll(el) {
    el.scrollTop(bottom(el));
}

function websocket(url) {
    if (url[0] == '/') url = location.origin + url;
    return new WebSocket(url.replace(/^http/, "ws").replace("thread", "socket"));
}

function threadLoad() {
    mainLoad();

    fmtTimes($("#right"), "d, t");

    var output = $("#msgs");
    MathJax.Hub.Register.StartupHook("End", function() { scroll(output); });

    var socket = websocket(location.href);
    socket.onmessage = function(e) {
        var m = JSON.parse(e.data);
        var username = create("div", "username light", m.Username + ":", []);
        var time = create("time", "light sans", "", []);
        time.attr("datetime", m.Time);
        fmtTime(time, "d, t");
        var body = create("div", "body" + (m.Tex ? " math" : ""), m.Body, []);
        var msg = create("div", "msg", "", [username, time, body]);
        if (m.Tex) MathJax.Hub.Queue(["Typeset", MathJax.Hub, body[0]]);
        var atBottom = output.scrollTop() == bottom(output)
        output.append(msg);
        if (atBottom) scroll(output);
    };

    var markdown = $("#markdown");
    var tex = $("#tex");
    var input = $("#input");
    input.keydown(function(e) {
        if (e.shiftKey && e.keyCode == 13) {
            var m = {
                "Body": input.val(),
                "Markdown": markdown.is(":checked"),
                "Tex": tex.is(":checked"),
            };
            input.val("");
            socket.send(JSON.stringify(m));
            e.preventDefault();
        }
    });
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

    var newthread = $("#newthread");
    $("#plusicon").click(function() {
        if (newthread.is(":visible")) {
            newthread.hide();
        } else {
            newthread.show();
            newthread.children(":first").focus();
        }
    });

    var left = $("#left");

    fmtTimes(left, "d<br/>t");

    var threads = $("#threads");
    threads.children().each(function() {
        var thread = $(this)
        var socket = websocket(thread.attr("href"));
        var lastmsg = thread.find(".lastmsg").first();
        var time = thread.find("time").first();
        socket.onmessage = function(e) {
            var m = JSON.parse(e.data);
            var d = new Date(m.Time);
            lastmsg.html(m.Username + ": ");
            lastmsg.append(create("span", m.Tex ? "math" : "", m.Body, []))
            time.attr("datetime", m.Time);
            fmtTime(time, "d<br/>t");
            if (m.Tex) MathJax.Hub.Queue(["Typeset", MathJax.Hub, lastmsg[0]]);
            thread.prependTo(threads).animate();
        };
    });

    var logo = $("#logo");
    var right = $("#right_wrap");
    logo.click(function() {
        if (left.css("left") == "0px") {
            left.animate({left: "-250px"}, 300);
            right.animate({marginLeft: "0px"}, 300);
        } else {
            left.animate({left: "0"}, 300, "swing");
            right.animate({marginLeft: "250px"}, 300);
        }
    });
}

function loginLoad() {
    var sw = $("#switch");
    var login = $("#login");
    var submit = $("#submit");
    var again = $("#again");
    sw.click(function() {
        if (again.is(":visible")) {
            submit.val("login");
            login.attr("action", "/login");
            again.hide();
            sw.val("need to register?");
        } else {
            submit.val("register");
            login.attr("action", "/register");
            again.show();
            sw.val("already have an account?");
        }
    });
}

var load = {"loginpage": loginLoad, "rootpage": rootLoad, "threadpage": threadLoad};

$(document).ready(function() { load[this.body.id](); });
