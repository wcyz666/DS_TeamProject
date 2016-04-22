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
        prevData = null,
        refreshButton = $("#btn-refresh"),
        notifArea = $("#notif-area"),
        utils = {
            trackUsage: function () {
                refreshButton.addClass("disabled");
                notifArea.empty();
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
                    utils.generateNotifications(data, utils.showNotifications);
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
            },

            generateNotifications : function (data, callback) {
                var i,
                    j = 0,
                    found = false;

                if (!prevData) {
                    prevData = data;
                    return;
                }
                // Positive notification
                for (i = 0; i < data.length; i++) {
                    found = false;
                    for (j = 0; j < prevData.length; j++) {
                        if (data[i].Name === prevData[j].Name) {
                            found = true;
                            break;
                        }
                    }
                    if (!found) {
                        callback && callback(new Notification("Node SuperNode [" + data[i].IP + ", " + data[i].Name + "] has" +
                            " joined the system.", true));
                    }
                }

                for (i = 0; i < prevData.length; i++) {
                    found = false;
                    for (j = 0; j < data.length; j++) {
                        if (data[j].Name === prevData[i].Name) {
                            found = true;
                            break;
                        }
                    }
                    if (!found) {
                        callback && callback(new Notification("Node SuperNode [" + prevData[i].IP + ", " + prevData[i].Name + "] has" +
                            " left the system.", false));
                    }
                }

                prevData = data;
            },

            showNotifications: function (notification) {
                notification.show();
            }
        };

    function Notification(content, isPositive) {
        this.isPositive = isPositive;
        this.content = content;
        this.HTML = $('<div class="alert" role="alert">' + this.content + '</div>')
            .addClass(this.isPositive? "alert-success" : "alert-danger");

    }
    Notification.prototype.show = function () {
        this.HTML.hide().appendTo(notifArea).slideDown();
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

