VIEWS.Signup = new View({
  Templates: {
    View: "#joinview",
  },
  Start: function() {
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
