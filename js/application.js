// Wait till the browser is ready to render the game (avoids glitches)
window.requestAnimationFrame(function () {
	var game = new GameManager(4, KeyboardInputManager, HTMLActuator, LocalStorageManager);

	window.move = function(d) {
		game.move(d);

		if(localStorage.gameState) {
			var state = JSON.parse(localStorage.gameState);
			if (state.over) return;

			setTimeout(function() {
				var e = document.createElement('script');
				e.type = 'text/javascript';
				e.async = true;
				e.src = 'http://localhost:8877?'+encodeURIComponent(JSON.stringify(state.grid));
				var s = document.getElementsByTagName('script')[0];
				s.parentNode.insertBefore(e, s);
			}, 0);
		}
	}

	window.move(0);
});
