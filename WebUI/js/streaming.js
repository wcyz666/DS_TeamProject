/**
 * Created by ASUA on 2016/4/23.
 */
$(document).ready(function() {

    var CONST = {
            NODE_URL : "",
            IS_STREAMER_URL : "/isStreamer/",
            RECEIVE_URL: "/receive/",
            SEND_URL: "/stream/",
            STOP_URL: "/stop/",
            GET_LOCAL_NAME_URL: "/getLocalName/",
            GET_TITLE_URL: "/getTitle/"
        },
        content = $('#chatroom-content'),
        text = $('#chatroom-text'),
        toggle = $('#streaming-toggle'),
        unviewMsg = 0,
        user = null,
        pageIsFocus = true;

    
    var myLib = (function(){

        var lastTime = new Date();

        return {
            getTime: function() {
                var now = new Date(),
                    wordsToHtml = '<p class="text-center small" id="datetime"></p>';
                if ((now - lastTime) / 1000 > 120 ) {
                    wordsToHtml = '<p class="text-center small" id="datetime">' + new Date().toLocaleString() + '</p>';
                }
                lastTime = now;
                return wordsToHtml;
            },
            getOtherWordsTemplate : function (words){
                var wordsToHtml = myLib.getTime();
                return wordsToHtml + '<div class="pull-left">' +
                    '</div><div class="media-body"><p class="words bubble-other">' + words.replace(/\n/g, "<br>") +
                    '</p></div><div class="clearfix"></div>';
            },
        };
    })();


    function User() {}


    User.prototype.updateUnread = function (msg) {

        var myWords = myLib.getOtherWordsTemplate(msg);
        content.append(myWords).animate({
            scrollTop:content[0].scrollHeight
        }, 500);
        if (!pageIsFocus) {
            ++unviewMsg;
            document.title = unviewMsg + " messages - Live Streaming";
        }
    };

    function Sender() {
    }

    Sender.prototype = new User();

    Sender.prototype.show = function () {
        var myWords = text.val();

        if (text.val().trim() === "") {
            return false;
        }

        text.val("");
        this.send(myWords);
        this.updateUnread(myWords);
    };

    Sender.prototype.send = function (msg) {
        $.get(CONST.SEND_URL + msg)
            .success(function () {
                console.log("Streaming out new message");
            })
            .error(function (e) {
                console.log(e);
            })
    };

    function Receiver(interval) {
        this.interval = interval;
        this.intervalHanler = null;
    }

    Receiver.prototype = new User();

    Receiver.prototype.show = function (msg) {
        this.updateUnread(msg);
    };

    Receiver.prototype.isStarted = false;
    
    Receiver.prototype.start = function () {
        this.isStarted = true;
        this.intervalHanler = setInterval(function () {
            $.ajax({
                url: CONST.NODE_URL + CONST.RECEIVE_URL,
                jsonp: "callback",
                dataType: "jsonp"
            }).success(function (data) {
                if (data.msg.length > 0) {
                    if (data.msg === "Control Message: Video begins") {
                        $('#video-panel').appendTo(content).removeClass("hidden");
                    } else {
                        user.show(data.msg);
                    }
                }
            }).error(function (data) {
                console.log(data);
            });

        }, this.interval);
    };

    Receiver.prototype.stop = function () {
        if (this.isStarted) {
            clearInterval(this.intervalHanler);
        }
        this.intervalHanler = null;
        this.isStarted = false;
    };

    (function (){

        function eventBinding() {

            $(document).keydown(function(event){
                if (event.keyCode == 13 || event.keyCode == 108) {
                    if (event.shiftKey) {
                        $('#sendMsg').click();
                        return false;
                    }
                }
            });
            $("#game").on('click', function(event){
                event.stopPropagation();

                switch (event.target.id){
                    case "sendVideo-span":
                    case "sendVideo":
                        $('#video-panel').appendTo(content).removeClass("hidden");
                        user.send("Control Message: Video begins");
                        break;
                    case "sendMsg":
                    case "sendMsg-span":
                        user.show();
                        break;
                }
            });



            $(window).focus(function(){
                unviewMsg = 0;
                pageIsFocus = true;
                document.title = "Live Streaming";
            });

            $(window).blur(function(){
                pageIsFocus = false;
            });
        }

        function initElement() {
            var btn2 = document.getElementById('reload-iframe-btn'),
                btn3 = document.getElementById('start-join-btn');
            
            var getTitleHandler = setInterval((function () {
                var titleUpdated = false;

                return function () {
                    $.ajax({
                        url: CONST.GET_TITLE_URL,
                        jsonp: "callback",
                        dataType: "jsonp"
                    }).success(function (data) {
                        if (titleUpdated) {
                            clearInterval(getTitleHandler);
                        }
                        if (data.title.length > 0 && !titleUpdated) {
                            var streamTitle = $('#streaming-title');
                            streamTitle.text(streamTitle.text() + " " + data.title);
                            titleUpdated = true;
                            clearInterval(getTitleHandler);
                        }
                    });
                };
            })(), 100);




            $.ajax({
                url: CONST.IS_STREAMER_URL,
                jsonp: "callback",
                dataType: "jsonp"
            }).success(function (data) {
                if (data.isStreamer) {
                    user = new Sender();
                    $('.panel-footer').removeClass("hidden");
                    text.height($('#sendMsg').height());

                    toggle.on('click', function () {
                        $.get(CONST.STOP_URL)
                            .success(function () {
                                console.log("Streaming Stop");
                            })
                            .error(function (e) {
                                console.log(e);
                            });
                    });

                } else {
                    user = new Receiver(1000);
                    user.start();

                    toggle.on('click', function () {
                        if (user.isStarted) {
                            user.stop();
                            $(this).text("Restart").removeClass('btn-warning').addClass("btn-success");
                        } else {
                            user.start();
                            $(this).text("Stop").addClass('btn-warning').removeClass("btn-success");
                        }

                    });
                }
            }).error(function (e) {
                console.log(e);
            });


            btn2.onclick = function () {
                var frame = document.getElementById("iframe");
                frame.contentWindow.postMessage({command: 'stop'}, '*');
            };

            btn3.onclick = function () {
                $.ajax({
                    url: CONST.GET_LOCAL_NAME_URL
                }).success(function (data) {
                    console.log(data);
                    var group = data;
                    var frame = document.getElementById("iframe");
                    frame.contentWindow.postMessage({command: 'start', groupid: group}, '*');
                }).error(function (data) {
                    console.log(data);
                })
            };
        }

        return {
            init : function () {
                eventBinding();
                initElement();
            }
        };
    })().init();
});