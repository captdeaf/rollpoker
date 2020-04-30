VIEWS.Menu = new View({
  Templates: {
    Menu: "#menucontents",
  },
  Start: function(el) {
    el.html(this.T.Menu({game: Game.data, player: Player.info}));
  },
  OnClick: {
    "asknotifications": function() {
      Notification.requestPermission();
    },
    "endgame": function() {
      if (confirm("Are you sure you want to end the game?")) {
        RollPoker.SendCommand("EndGame", {});
      }
    },
  },
  OnChange: {
    "MenuDisplayName": function(evt) {
      var ndn = evt.target.value;
      if (!ndn.match(/\S/)) {
        evt.target.value = Player.info.DisplayName;
      } else {
        RollPoker.SendCommand("PlayerUpdate", {
          DisplayName: evt.target.value
        });
      }
    }
  },
  Update: function(data) {
  },
});
