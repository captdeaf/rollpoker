VIEWS.Menu = new View({
  Templates: {
    Menu: "#menucontents",
    Settings: "#editsettingsview",
    Hosts: "#managehostsview",
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
    "editsettings": function() {
      var dlg = $('#dialog');
      dlg.find("#dialogcontents").html(this.T.Settings());
      dlg.show();
    },
    "closedialog": function() {
      $('#dialog').hide();
    },
    "edithosts": function() {
      var dlg = $('#dialog');
      dlg.find("#dialogcontents").html(this.T.Hosts());
      dlg.show();
    },
    "promote": function(evt) {
      var playerid = evt.target.getAttribute('data');
      RollPoker.SendCommand("Promote", {"PlayerId": playerid});
      var me = this;
      setTimeout(function() {
        $("#dialogcontents").html(me.T.Hosts());
      }, 1000);
    },
    "demote": function(evt) {
      var playerid = evt.target.getAttribute('data');
      RollPoker.SendCommand("Demote", {"PlayerId": playerid});
      var me = this;
      setTimeout(function() {
        $("#dialogcontents").html(me.T.Hosts());
      }, 1000);
    },
  },
  OnSubmit: {
    "gamesettings": function() {
      var data = {}
      _.each($('#gamesettings').serializeArray(), function(fd) {
        data[fd.name] = "" + fd.value;
      });
      RollPoker.SendCommand("UpdateSettings", data);
      var me = this;
      setTimeout(function() {
        $("#dialogcontents").html(me.T.Settings());
      }, 2000);
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
    console.log("Menu Update called?");
  },
});
