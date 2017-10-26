
$( document ).ready(function() {
    console.log( "ready!" );
});

$( '#openStrava' ).on('click', function(event) {
	event.preventDefault();
	console.log('open '+authUri+' '+window.parent.location.href);
	var directUri = String(window.parent.location.href).replace('#!/', '');
	var ix = directUri.indexOf('?');
	if (ix>=0) { directUri = directUri.substring(0,ix); }
	window.parent.location.href = authUri+'redirect_uri='+encodeURIComponent(directUri /*+'/auth_callback'*/);
	return false;
});

function doSync(event) {
	event.preventDefault();
	console.log('sync ');
	$.post('./ui/api/sync', {})
	.done(function () {
		console.log('Done sync');
	})
	.fail(function() {
		console.log('Error requesting sync');
	});
	return false;
}