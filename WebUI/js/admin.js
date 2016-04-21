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

        function eventBinding() {
            $("")
        }

        function initElement() {
            $('input[type=checkbox]').bootstrapSwitch();
        }

        return {
            init: function() {
                eventBinding();
                initElement();
            }
        }
    })().init();
});

