/**
 * Created by ASUA on 2016/4/23.
 */
$(document).ready(function() {

    var CONST = {
            NODE_URL : "http://localhost:9999",
            IS_STREAMER_URL : "/isStreamer/"
        },
        content = $('#chatroom-content'),
        text = $('#chatroom-text'),
        unviewMsg = 0,
        pageIsFocus = true;

    var myLib = (function(){

        var lastTime = new Date(),
            width = $("#game").width() * 0.3;

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
            getWordsTemplate : function (words){
                var wordsToHtml = myLib.getTime();
                return wordsToHtml + '<div class="pull-right"><img class="media-object" width="48" src="/avatar/' + userID + '.png" alt="avatar">'+
                    '</div><div class="media-body word-content pull-right"><p class="bubble-self words col-xs-12">' + words.replace(/\n/g, "<br>") +
                    '</p></div><div class="clearfix"></div>';
            },


            msgCallback: function() {
                if (!pageIsFocus) {
                    ++unviewMsg;
                    document.title = unviewMsg + " messages - Live Streaming";
                }
            }
        };
    })();

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
                        var myWords = myLib.getWordsTemplate(text.val());

                        if (text.val().trim() === "")
                            return false;

                        text.val("");
                        content.append(myWords).animate({
                            scrollTop:content[0].scrollHeight
                        }, 500);
                        myLib.msgCallback();
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
                    $('.panel-footer').removeClass("hidden");
                    text.height($('#sendMsg').height());
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