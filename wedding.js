$(document).ready(function() {
  $('#fullpage').fullpage({
	//Navigation
	menu: '#menu',
	anchors: ['welcome', 'weddingDay', 'gifts', 'rsvp'],
	navigation: false,
	
	verticalCentered: true,
	sectionsColor: ['', '#FFFFFF', '#FFFFF0', '#F5F5DC'],
  });
});