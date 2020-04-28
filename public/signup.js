VIEWS.Signup = new View({
  Templates: {
    View: "#joinview",
  },
  Start: function() {
    this.init();
    $('body').html(this.T.View({data: {Players: []}}));
  },
  Update: function(state) {
    $('body').html(this.T.View({data: state}));
  },
  OnClick: {
    "startgame": function(evt) {
      RollPoker.SendCommand("StartPoker");
    },
  },
  OnSubmit: {
    "signup": function(evt) {
      var dispname = $('#DisplayName').val();
      RollPoker.SendCommand("Register", {DisplayName: dispname});
    },
    "joingroup": function(evt) {
      var roompass = $('#RoomPass').val();
      RollPoker.SendCommand("JoinGroup", {RoomPass: roompass}, function() {
        window.location.reload(false);
      });
    },
  }
});
