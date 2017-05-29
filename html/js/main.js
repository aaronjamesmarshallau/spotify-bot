$(document).ready(function () {
	Spotify.init();

	Spotify.assertAuthStatus(Spotify.start,	Spotify.showLoginPage);
});
