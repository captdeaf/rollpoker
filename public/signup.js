VIEWS.Signup = new View({
  Templates: {
    View: "#joinview",
    PlayerList: "#playerlistview",
    PlayerJoin: "#playerjoinview",
  },
  Start: function() {
    $('#sizer').html(this.T.View({data: {Players: []}}));
  },
  Update: function(state) {
    $('#playerlist').html(this.T.PlayerList({data: state}));
    $('#invites').html(this.T.PlayerJoin({data: state}));
  },
  OnClick: {
    "startgame": function(evt) {
      RollPoker.SendCommand("StartPoker");
    },
    "cardsettings": function() {
      var menu = $("#menu");
      this.AddSubview("#menu", VIEWS.Menu);
      menu.show();
    },
    "closemenu": function() {
      this.RemoveSubview("#menu");
      $("#menu").hide();
    },
    "kick": function(evt) {
      var pid = evt.target.getAttribute("data");
      console.log("Kicking " + pid);
      RollPoker.SendCommand("KickPlayer", {"PlayerId": pid});
    },
  },
  OnSubmit: {
    "signup": function(evt) {
      RollPoker.SendCommand("Register", {});
    },
    "joingroup": function(evt) {
      var roompass = $('#RoomPass').val();
      var dispname = $('#DisplayName').val();
      RollPoker.SendCommand("JoinGroup", {DisplayName: dispname, RoomPass: roompass}, function() {
        window.location.reload(false);
      });
    },
  }
});
