VIEWS.Menu = new View({
  Templates: {
    Menu: "#menucontents",
  },
  Start: function(el) {
    console.log("Menu.start");
    el.html(this.T.Menu({game: Game.data, player: Player.info}));
  },
  OnClick: {
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
