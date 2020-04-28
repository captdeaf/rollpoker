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
    }
  },
  OnSubmit: {
    "joinup": function(evt) {
      var data = {}
      _.each($('#joingameform').serializeArray(), function(fd) {
        data[fd.name] = fd.value;
      });
      // TODO: Sanity check email and display name
      RollPoker.SendCommand("Join", data);
    }
  }
});
