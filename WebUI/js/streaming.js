/**
 * Created by ASUA on 2016/4/23.
 */
$(document).ready(function() {

    var CONST = {
            NODE_URL : "",
            IS_STREAMER_URL : "/isStreamer/",
            RECEIVE_URL: "/receive/",
            SEND_URL: "/stream/",
            STOP_URL: "/stop/"
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
    
    Receiver.prototype.start = function () {
        this.intervalHanler = setInterval(function () {
            $.ajax({
                url: CONST.NODE_URL + CONST.RECEIVE_URL,
                jsonp: "callback",
                dataType: "jsonp"
            }).success(function (data) {
                if (data.msg.length > 0) {
                    user.show(data.msg);
                }
            }).error(function (data) {
                console.log(data);
            });

        }, this.interval);
    };

    Receiver.prototype.stop = function () {
        if (this.intervalHanler) {
            clearInterval(this.intervalHanler);
        }
        this.intervalHanler = null;
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
            $.ajax({
                url: CONST.NODE_URL + CONST.IS_STREAMER_URL,
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
                        user.stop();
                    });
                }
            }).error(function (e) {
                console.log(e);
            })
        }

        return {
            init : function () {
                eventBinding();
                initElement();
            }
        };
    })().init();
});