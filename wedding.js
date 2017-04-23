adult_entry = "<div class=\"col-xs-12\">\r\n<label>Adult Guest<\/label>\r\n<\/div>\t\t\r\n<div class=\"col-xs-4\">\r\n<div class=\"form-group\">\r\n<label>First Name<\/label>\r\n<input type=\"\" class=\"form-control\" id=\"\" placeholder=\"\">\r\n<\/div>\r\n<\/div>\r\n<div class=\"col-xs-4\">\r\n<div class=\"form-group\">\r\n<label>Last Name<\/label>\r\n<input type=\"\" class=\"form-control\" id=\"\" placeholder=\"\">\r\n<\/div>\r\n<\/div>\r\n<div class=\"col-xs-4\">\r\n<div class=\"form-group\">\r\n<label>Entr\u00E9e<\/label>\r\n<select class=\"form-control\">\r\n<option>Beef<\/option>\r\n<option>Cod<\/option>\r\n<option>Veggie<\/option>\r\n<\/select>\r\n<\/div>\r\n<\/div>"
child_entry = "<div class=\"col-xs-12\">\r\n<label>Child Guest<\/label>\r\n<\/div>\t\t\r\n<div class=\"col-xs-4\">\r\n<div class=\"form-group\">\r\n<label>First Name<\/label>\r\n<input type=\"\" class=\"form-control\" id=\"\" placeholder=\"\">\r\n<\/div>\r\n<\/div>\r\n<div class=\"col-xs-4\">\r\n<div class=\"form-group\">\r\n<label>Last Name<\/label>\r\n<input type=\"\" class=\"form-control\" id=\"\" placeholder=\"\">\r\n<\/div>\r\n<\/div>\r\n<div class=\"col-xs-4\">\r\n<div class=\"form-group\">\r\n<label>Entr\u00E9e<\/label>\r\n<select class=\"form-control\">\r\n<option>Beef<\/option>\r\n<option>Cod<\/option>\r\n<option>Veggie<\/option>\r\n<\/select>\r\n<\/div>\r\n<\/div>"

$(document).ready(function() {
  $('#fullpage').fullpage({
	//Navigation
	menu: '#menu',
	anchors: ['welcome', 'weddingDay', 'gifts', 'rsvp'],
	navigation: false,
	
	verticalCentered: true,
	sectionsColor: ['', '#FFFFFF', '#FFFFF0', '#F5F5DC'],
  });
  $("#guests-adult-entry").empty();
  $("#guests-child-entry").empty();

  $("#rsvp-form #guest-adult").change(function() {
	  var adults_count = $("#guest-adult").val();
	  $("#guests-adult-entry").empty();
	  for (var i = 0; i < adults_count; i++) {
		  $("#guests-adult-entry").append(adult_entry);
	  }  
  })
  
  $("#rsvp-form #guest-child").change(function() {
	  var adults_count = $("#guest-child").val();
	  $("#guests-child-entry").empty();
	  for (var i = 0; i < adults_count; i++) {
		  $("#guests-child-entry").append(child_entry);
	  }  
  })
});