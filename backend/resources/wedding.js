$(document).ready(function() {
    $('#fullpage').fullpage({
        //Navigation
        menu: '#menu',
        anchors: ['welcome', 'weddingDay', 'gifts', 'rsvp'],
        navigation: false,
        scrollOverflow: true,

        verticalCentered: true,
        sectionsColor: ['', '#FFFFFF', '#FFFFF0', '#F5F5DC'],
    });

    var adult_entry = $("#guests-adult-entry .entry")
    var child_entry = $("#guests-child-entry .entry")
    var adult_entry_header = $("#guests-adult-entry .entry-header")
    var child_entry_header = $("#guests-child-entry .entry-header")
    $("#guests-adult-entry").empty();
    $("#guests-child-entry").empty();

    $("#rsvp-form #guest-adult").change(function() {
        var adults_count = $("#guest-adult").val();
        $("#guests-adult-entry").empty();
        if (adults_count > 0) {
            $("#guests-adult-entry").append(adult_entry_header)
        }
        for (var i = 0; i < adults_count; i++) {
            var adult_entry_clone = adult_entry.clone()
            adult_entry_clone.find("[name='guestAdultFirstName']").attr("name", "guestAdultFirstName" + i)
            adult_entry_clone.find("[name='guestAdultLastName']").attr("name", "guestAdultLastName" + i)
            adult_entry_clone.find("[name='guestAdultEntree']").attr("name", "guestAdultEntree" + i)
            $("#guests-adult-entry").append(adult_entry_clone);
        }
		$.fn.fullpage.reBuild();
    })

    $("#rsvp-form #guest-child").change(function() {
        var children_count = $("#guest-child").val();
        $("#guests-child-entry").empty();
        if (children_count > 0) {
            $("#guests-child-entry").append(child_entry_header)
        }
        for (var i = 0; i < children_count; i++) {
            var child_entry_clone = child_entry.clone()
            child_entry_clone.find("[name='guestChildFirstName']").attr("name", "guestChildFirstName" + i)
            child_entry_clone.find("[name='guestChildLastName']").attr("name", "guestChildLastName" + i)
            child_entry_clone.find("[name='guestChildEntree']").attr("name", "guestChildEntree" + i)
            $("#guests-child-entry").append(child_entry_clone);
        }
		$.fn.fullpage.reBuild();
    })

    $("#rsvp-form").submit(function(e) {
        e.preventDefault();
        $.post("/rsvp", $("#rsvp-form").serialize(), function(response) {
			response = JSON.parse(response)
			if (response["success"]) {
				$("#rsvp-container").empty();
				$("#rsvp-container").append("<div class=\"text-center\"><h2>" + response["message"] + "</h2></div>");
			} else {
				$("#error").text(response["message"]);
			}
        });
    })
});