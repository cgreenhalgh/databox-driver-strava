
$( document ).ready(function() {
    console.log( "ready!" );
});

$( '#openStrava' ).on('click', function(event) {
	event.preventDefault();
	console.log('open '+authUri+' '+window.parent.location.href);
	window.parent.location.href = authUri+'redirect_uri='+encodeURIComponent(window.parent.location.href /*+'/auth_callback'*/);
	return false;
});