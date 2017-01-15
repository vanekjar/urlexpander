$(document).ready(function () {
    $("#submit-btn").click(function () {
        // hide all previous results
        $("#result *").removeClass("in");

        var url = $("#url").val();
        if (!url) return;

        var $btn = $(this).button('loading');

        $.ajax("/api/expand?url=" + url)
            .done(function (data) {
                data = JSON.parse(data);

                $("#result-success a.result-url").attr("href", data.expanded)
                $("#result-success a.result-url").html(data.expanded)
                $("#result-success").collapse("show");
            })
            .fail(function (jqXHR) {
                var data = JSON.parse(jqXHR.responseText);
                $("#result-error .error-desc").html(data.error);
                $("#result-error").collapse("show")
            })
            .always(function () {
                $btn.button('reset');
            });
    });

    // submit form by pressing ENTER
    $("input#url").keypress(function(e) {
        if (e.keyCode == 13) {
            $("#submit-btn").click();
        }
    });
});