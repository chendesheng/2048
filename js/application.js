// Wait till the browser is ready to render the game (avoids glitches)
window.requestAnimationFrame(function () {
	var game = new GameManager(4, KeyboardInputManager, HTMLActuator, LocalStorageManager);

	window.move = function(d) {
		if (game.isGameTerminated()) {
			game.keepPlaying = true;
			setTimeout(function() {
				document.querySelector(".game-message").classList.remove("game-won")
			}, 100);
		}

		game.move(d);

		if(localStorage.gameState) {
			var state = JSON.parse(localStorage.gameState);
			if (state.over) return;

			var e = document.createElement('script');
			e.type = 'text/javascript';
			e.async = true;
			e.src = 'http://localhost:8877?'+encodeURIComponent(JSON.stringify(state.grid));
			var s = document.getElementsByTagName('script')[0];
			s.parentNode.insertBefore(e, s);
		}
	}

	window.move(0);
});
