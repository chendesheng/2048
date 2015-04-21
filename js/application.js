// Wait till the browser is ready to render the game (avoids glitches)
window.requestAnimationFrame(function () {
	var game = new GameManager(4, KeyboardInputManager, HTMLActuator, LocalStorageManager);

	var onemove = function() {
		if(localStorage.gameState) {
			var state = JSON.parse(localStorage.gameState);
			if (state.over) return;

			game.move(Math.ceil(Math.random()*10)%4);
			setTimeout(onemove, 100);
		}
	}

	onemove();
});
