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
    return function() { el.scrollTop(bottom(el)); };
}

function websocket(url) {
    if (url[0] == '/') url = location.origin + url;
    return new WebSocket(url.replace(/^http/, "ws").replace("thread", "socket"));
}

function mathjax(el) {
    MathJax.Hub.Queue(["Typeset", MathJax.Hub, el[0]]);
}

function threadLoad() {
    mainLoad();

    var msgs = $("#msgs");
    MathJax.Hub.Register.StartupHook("End", scroll(msgs));

    fmtTimes($("#right"), "d, t");

    var bodySwitch = function() {
        $(this).parent().find(".body").toggleClass("hide");
    }

    msgs.find(".raw").click(bodySwitch);

    var socket = websocket(location.href);
    socket.onmessage = function(e) {
        var m = JSON.parse(e.data);
        var username = create("div", "username light", m.Username + ":", []);
        var fmtBody = create("div", "body" + (m.Tex ? " math" : ""), m.FmtBody, []);

        var time = create("time", "superlight sans", "", []);
        time.attr("datetime", m.Time);
        fmtTime(time, "d, t");

        var msg = create("div", "msg", "", [username, time, fmtBody]);

        if (m.Markdown || m.Tex) {
            var rawBody = create("div", "body hide", m.RawBody, []);
            var rawImg = create("img", "", "", []);
            rawImg.attr("src", "/static/dark-arrow.png");
            var raw = create("div", "raw pointer", "", [rawImg]);
            raw.click(bodySwitch);
            fmtBody.before(raw, rawBody);
        }

        var atBottom = msgs.scrollTop() == bottom(msgs)
        if (m.Tex) {
            mathjax(fmtBody);
            msgs.append(msg);
            if (atBottom) MathJax.Hub.Queue(scroll(msgs));
        } else {
            msgs.append(msg);
            if (atBottom) scroll(msgs)();
        }
    };

    var markdown = $("#markdown");
    var tex = $("#tex");
    var input = $("#input");
    var send = $("#send");
    var preview = $("#preview");
    var previewContent = $("#preview_content");
    var down = $("#downicon");

    var hidePreview = function() {
        previewContent.html("");
        down.hide(200);
        previewContent.animate({bottom: "30px"}, 150);
        var atBottom = msgs.scrollTop() == bottom(msgs);
        msgs.animate({
            bottom: "140px",
            scrollTop: msgs.scrollTop() - 100,
        }, 150, function() {
            if (atBottom) MathJax.Hub.Queue(scroll(msgs));
        });
    }

    var onsend = function() {
        var m = {
            "RawBody": input.val(),
            "Markdown": markdown.is(":checked"),
            "Tex": tex.is(":checked"),
        };
        input.val("");
        socket.send(JSON.stringify(m));
        if (previewContent.css("bottom") == "120px") {
            hidePreview();
        }
    };

    input.keydown(function(e) {
        if (e.shiftKey && e.keyCode == 13) {
            onsend();
            e.preventDefault();
        }
    });
    send.click(onsend);

    preview.click(function() {
        var m = {
            "body": input.val(),
            "markdown": markdown.is(":checked") ? "md" : "",
            "tex": tex.is(":checked") ? "tex" : "",
        };
        previewContent.load("/preview", m, function() {
            if (tex.is(":checked")) mathjax(previewContent);
            if (previewContent.css("bottom") == "30px") {
                down.show(200);
                previewContent.animate({bottom: "120px"}, 150);
                msgs.animate({
                    bottom: "240px",
                    scrollTop: msgs.scrollTop() + 100,
                }, 150);
            }
        });
    });

    down.click(hidePreview);
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
    $("#addthread").click(function() {
        if (newthread.is(":visible")) {
            newthread.slideUp(150);
        } else {
            newthread.slideDown(150);
            newthread.children(":first").focus();
        }
    });

    var left = $("#left");

    fmtTimes(left, "d<br>t");

    var threads = $("#threads");
    threads.children().each(function() {
        var thread = $(this);
        var socket = websocket(thread.attr("href"));
        var lastmsg = thread.find(".lastmsg").first();
        var time = thread.find("time").first();
        socket.onmessage = function(e) {
            var m = JSON.parse(e.data);
            lastmsg.html(m.Username + ": " + m.FmtBody);

            var d = new Date(m.Time);
            time.attr("datetime", m.Time);
            fmtTime(time, "d<br>t");

            thread.prependTo(threads);
        };
    });

    var logo = $("#logo");
    var right = $("#right_wrap");
    logo.click(function() {
        if (left.css("left") == "0px") {
            left.animate({left: "-300px"}, 150);
            right.animate({left: "0px"}, 150);
        } else {
            left.animate({left: "0"}, 150);
            right.animate({left: "300px"}, 150);
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
