/**
 * Created by ASUA on 2016/4/17.
 */



$(document).ready(function () {

    var CONST = {
            NODE_URL : "http://localhost:9999",
            TRACK_LOAD_URL : "/load/",
            TOPO_EDGE_LENGTH : 2
        },
        refreshInterval = -1,
        refreshHandler = null,
        prevData = null,
        refreshButton = $("#btn-refresh"),
        notifArea = $("#notif-area"),
        graphArea;
        utils = {
            trackUsage: function () {
                var isChanged;

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
                        html += "<tr><td>" + data[i].IP + "</td><td>"
                            + data[i].Name + "</td><td>"
                            + data[i].ChildCount + "</td><td>"
                            + '<button class="btn btn-sm btn-info"><span class="glyphicon glyphicon-minus" aria-hidden="true"></span>' +
                            '</button></td></tr>';
                    }
                    $("#table-body").html(html);
                    isChanged = utils.generateNotifications(data, utils.showNotifications);
                    refreshButton.removeClass("disabled");

                    if (isChanged) {
                        utils.generateGraph(data, utils.redraw);
                    }

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
                    found = false,
                    isChanged = false;

                if (!prevData) {
                    prevData = data;
                    return true;
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
                        isChanged = true;
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
                        isChanged = true;
                    }
                }

                prevData = data;
                return isChanged;
            },

            showNotifications: function (notification) {
                notification.show();
            },

            generateGraph: function (rawData, callback) {
                var i,
                    graph = {
                        "nodes": [],
                        "edges": []
                    },
                    len,
                    node,
                    edge,
                    coordinate;

                for (i = 0, len = rawData.length; i < len; i++) {
                    node = {
                        id : "n" + i,
                        size : 3
                    };

                    node.label = "IP: " + rawData[i].IP + ", ChildCount: " + rawData[i].ChildCount;
                    coordinate = utils.getCoordinates(CONST.TOPO_EDGE_LENGTH, i + 1, len);
                    node.x = coordinate.x;
                    node.y = coordinate.y;

                    edge = {
                        id : "e" + i,
                        source : "n" + i,
                        target : "n" + ((i + 1) % len)
                    };

                    graph.nodes.push(node);
                    graph.edges.push(edge);
                }

                callback && callback(graph);
            },

            redraw: function (data) {
                if (graphArea) {
                    graphArea.graph.clear();
                    graphArea.graph.read(data);
                    graphArea.refresh();
                } else {
                    graphArea = new sigma({
                        graph: data,
                        container: 'topo-area',
                        settings: {
                            defaultNodeColor: '#ec5148'
                        }
                    });
                }

            },

            getCoordinates: function (edgeLength, order, edgeCount) {
                return {
                    "x" : edgeLength * Math.cos(Math.PI / 2 + (order - 1 ) * Math.PI * 2 / edgeCount),
                    "y" : edgeLength * Math.sin(Math.PI / 2 + (order - 1 ) * Math.PI * 2 / edgeCount)
                };
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
            utils.trackUsage();

        }


        return {
            init: function() {
                eventBinding();
                initElement();

            }
        }
    })().init();
});

