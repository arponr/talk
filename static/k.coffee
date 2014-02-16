create = (tag) -> $(document.createElement(tag))
bottom = (el) -> el[0].scrollHeight - el[0].offsetHeight
scroll = (el) -> el.scrollTop bottom(el)
mathjax = (el) -> MathJax.Hub.Queue ["Typeset", MathJax.Hub, el[0]]
prezero = (x) -> return if x < 10 then "0" + x else x

printDate = (t) ->
    y = t.getFullYear()
    m = prezero t.getMonth() + 1
    d = prezero t.getDate()
    return "#{y}-#{m}-#{d}"

printTime = (t) ->
    h = t.getHours()
    m = prezero t.getMinutes()
    p = "am"
    if h == 0
        h = 12
    else if h > 12
        h -= 12
        p = "pm"
    return "#{h}:#{m}#{p}"

fmtTime = (time, fmt) ->
    dt = new Date time.attr("datetime")
    d = printDate(dt)
    t = printTime(dt)
    time.html fmt.replace("d", d).replace("t", t)

fmtTimes = (el, fmt) -> el.find("time").each(-> fmtTime($(this), fmt))

websocket = (url) ->
    if url[0] == '/'
        url = location.origin + url
    return new WebSocket url.replace(/^http/, "ws").replace("thread", "socket")

threadLoad = ->
    mainLoad()

    msgs = $("#msgs")
    MathJax.Hub.Register.StartupHook "End", -> scroll(msgs)

    fmt = "d, t"
    fmtTimes $("#right"), fmt

    bodySwitch = ->
        $(this).closest(".msg").find(".body").toggle()

    msgs.find(".body_switch").click(bodySwitch)

    socket = websocket location.href
    socket.onmessage = (e) ->
        m = JSON.parse e.data
        username = create("div").addClass("username").html("#{m.Username}")
        fmtBody = create("div")
            .addClass("body fmt#{" math" if m.Tex}")
            .html(m.FmtBody)
        time = create("time").attr(datetime: m.Time)
        fmtTime time, fmt
        aside = create("div").addClass("aside").append(time)
        msg = create("div").addClass("msg").append([aside, username, fmtBody])
        if m.Markdown or m.Tex
            rawBody = create("div").addClass("body raw").html(m.RawBody)
            fmtBody.before(rawBody)
            time.addClass("body_switch")
                .click(bodySwitch)
        if m.Tex
            mathjax fmtBody
        atBottom = msgs.scrollTop() == bottom(msgs)
        msgs.append msg
        if atBottom
            MathJax.Hub.Queue(-> scroll(msgs))

    markdown = $("#markdown")
    tex = $("#tex")
    input = $("#input")
    send = $("#send")
    preview = $("#preview")
    previewContent = $("#preview_content")
    down = $("#downicon")

    hidePreview = ->
        atBottom = msgs.scrollTop() == bottom(msgs);
        previewContent.html ""
        down.hide 150
        previewContent.animate {bottom: "30px"}, 150
        msgs.animate {bottom: "140px", scrollTop: msgs.scrollTop() - 100}, 150,
            MathJax.Hub.Queue(-> scroll(msgs) if atBottom)

    onsend = ->
        m =
            "RawBody": input.val()
            "Markdown": markdown.is(":checked")
            "Tex": tex.is(":checked")
        input.val ""
        socket.send JSON.stringify(m)
        if previewContent.css("bottom") == "120px"
            hidePreview()

    input.keydown (e) ->
        if e.shiftKey and e.keyCode == 13
            onsend()
            e.preventDefault()

    send.click onsend

    preview.click ->
        mdCheck = markdown.is ":checked"
        texCheck = tex.is ":checked"
        m =
            "raw": input.val()
            "markdown": if mdCheck then "md" else ""
            "tex": if texCheck then "tex" else ""
        previewContent.load "/preview", m, ->
            if texCheck
                mathjax(previewContent)
            if previewContent.css("bottom") == "30px"
                down.show 150
                previewContent.animate {bottom: "120px"}, 150
                msgs.animate {bottom: "240px", scrollTop: msgs.scrollTop() + 100}, 150

    down.click hidePreview

rootLoad = ->
    mainLoad()

mainLoad = ->
    MathJax.Hub.Config
        tex2jax:
            inlineMath: [['$','$']]
            processClass: "math"
            ignoreClass: "nomath"
        "HTML-CSS":
            scale: 95
            availableFonts: []
            webFont: "Gyre-Termes"

    newthread = $("#newthread")
    $("#addthread").click ->
        if newthread.is ":visible"
            newthread.slideUp 150
        else
            newthread.slideDown 150
            newthread.children(":first").focus()

    left = $("#left")

    fmt = "d<br>t"
    fmtTimes left, fmt

    threads = $("#threads")
    threads.children().each ->
        thread = $(this)
        socket = websocket thread.attr("href")
        lastmsg = thread.find(".lastmsg").first()
        time = thread.find("time").first()
        socket.onmessage = (e) ->
            m = JSON.parse e.data
            lastmsg.html "#{m.Username}: #{m.FmtBody}"
            time.attr datetime: m.Time
            fmtTime time, fmt
            thread.prependTo threads

    logo = $("#logo")
    right = $("#right_wrap")
    logo.click ->
        if left.css("left") == "0px"
            left.animate  {left: "-260px"}, 150
            right.animate {left: "0px"},    150
        else
            left.animate  {left: "0"},     150
            right.animate {left: "260px"}, 150

loginLoad = ->
    sw = $("#switch")
    login = $("#login")
    submit = $("#submit")
    again = $("#again")
    sw.click ->
        if again.is ":visible"
            submit.val "login"
            login.attr action: "/login"
            sw.val "need to register?"
            again.hide()
        else
            submit.val "register"
            login.attr action: "/register"
            sw.val "already have an account?"
            again.show()

load =
    "loginpage":  loginLoad
    "rootpage":   rootLoad
    "threadpage": threadLoad

jQuery -> load[this.body.id]()
