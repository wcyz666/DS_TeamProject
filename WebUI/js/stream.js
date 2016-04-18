/**
 * Created by ASUA on 2016/4/17.
 */



$(document).ready(function () {

    var CONST = {
        SUPERNODE_URL : "http://52.90.181.29:9999",
        GET_LIVE_LIST_URL : "/allPrograms/",
        JOIN_LIVE_URL : "/join/"
    };

    (function () {

        function eventBinding(){
            $('#streamer, #viewer').on("click", function (event) {
                console.log("#" + event.target.id + "-Modal");
                $("#" + event.target.id + "-Modal").modal();
            });

            $("#new-live").on('submit', function (event) {
                event.preventDefault();
                var url = $(this).attr("action") + $('#live-title').val();
                window.location.href = url;
                return false;
            });

            $("#viewer").on('click', function () {
                $.ajax({
                    url: CONST.SUPERNODE_URL + CONST.GET_LIVE_LIST_URL,
                    jsonp: "callback",
                    dataType: "jsonp"
                })
                    .success(function (data) {
                        var html = "";
                        for (var key in data) {
                            html += "<option value='" + key + "'>" + data[key] + "</option>"
                        }
                        $('#live-list').html(html);
                    })
                    .error(function (data) {
                        console.log(data);
                    })
            });

            $("#live-join").on('submit', function (event) {
                event.preventDefault();
                var url = CONST.SUPERNODE_URL + CONST.JOIN_LIVE_URL
                    + $(this).find("select option:selected").val();
                window.location.href = url;
                return false;
            });
        }


        return {
            init: function() {
                eventBinding();
                $('#fullpage').fullpage({
                    menu: '#menu',
                    sectionsColor: ['#2B2B2B', '#C63D0F', '#F6F6F6', '#FFE200'],
                    anchors: ['firstPage', 'secondPage', '3rdPage', '4thpage'],
                    scrollingSpeed: 1000,
                    css3: true,
                    easingcss3: 'cubic-bezier(0.175, 0.685, 0.320, 1.275)',
                    slidesNavigation: true,
                    slidesNavPosition: 'bottom',
                    fixedElements: '.modal'
                });
            }
        }
    })().init();
});

