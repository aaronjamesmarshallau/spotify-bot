$(document).ready(function () {
	Spotify.init();

	Spotify.registerClient(function () {
		Spotify.start();
	});
});
