/**
 * Created by ASUA on 2016/4/17.
 */



$(document).ready(function () {

    var CONST = {
        NODE_URL : "http://localhost:9999",
        TRACK_LOAD_URL : "/load/"
    },
        refreshInterval = -1,
        refreshHandler = null,
        refreshButton = $("#btn-refresh"),
        utils = {
            trackUsage: function () {
                refreshButton.addClass("disabled");
                $.ajax({
                    url: CONST.NODE_URL + CONST.TRACK_LOAD_URL,
                    jsonp: "callback",
                    dataType: "jsonp"
                }).success(function (data) {
                    var html = "",
                        i = 0;
                    for (; i < data.length; i++) {
                        html += "<tr><th>" + data[i].IP + "</th><th>"
                            + data[i].Name + "</th><th>"
                            + data[i].ChildCount + "</th></tr>";
                    }
                    $("#table-body").html(html);
                    refreshButton.removeClass("disabled");
                }).error(function (data) {
                    console.log(data);
                });
            },

            startAutoRefresh: function (newInterval) {
                if (refreshHandler) {
                    clearInterval(refreshHandler);
                }
                refreshHandler = setInterval(function () {
                    utils.trackUsage();
                }, newInterval);
            }

        };


    (function () {

        function eventBinding() {
            refreshButton.on("click", function () {
                utils.trackUsage();
            });

            $(".btn-group").click(function (event) {
                var $target = $(event.target).find('input');
                switch ($target.attr("id")) {
                    case "option0":
                        if (refreshInterval != -1) {
                            clearInterval(refreshHandler);
                            refreshHandler = null;
                        }
                        refreshInterval = -1;
                        return;
                    case "option1":
                        refreshInterval = 10 * 1000;
                        break;
                    case "option2":
                        refreshInterval = 30 * 1000;
                        break;
                    case "option3":
                        refreshInterval = 60 * 1000;
                }
                utils.startAutoRefresh(refreshInterval);
            });
        }

        function initElement() {
            utils.trackUsage()
        }


        return {
            init: function() {
                eventBinding();
                initElement();
            }
        }
    })().init();
});

