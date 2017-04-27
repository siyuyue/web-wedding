$(document).ready(function() {
  $('#fullpage').fullpage({
	//Navigation
	menu: '#menu',
	anchors: ['welcome', 'weddingDay', 'gifts', 'rsvp'],
	navigation: false,
	
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
		  $("#guests-adult-entry").append(adult_entry.clone());
	  }  
  })
  
  $("#rsvp-form #guest-child").change(function() {
	  var children_count = $("#guest-child").val();
	  $("#guests-child-entry").empty();
	  if (children_count > 0) {
		  $("#guests-child-entry").append(child_entry_header)
	  }
	  for (var i = 0; i < children_count; i++) {
		  $("#guests-child-entry").append(child_entry.clone());
	  }  
  })
  
  $("#rsvp-form").submit(function() {
	$.post("/rsvp", function(response) {
	  alert(response);
	});
  })
});