var POKER = {
  Setup: function() {
    // First make sure we have our game name.
    var m = document.location.pathname.match(/table\/(\w+)$/);
    if (m) {
      POKER.NAME = m[1];
    } else {
      return;
    }
    var m = document.location.search.match(/\?id=(\w+)\&key=(\w+)$/);
    if (m) {
      POKER.PLAYER_ID = m[1];
      POKER.PLAYER_KEY = m[2];
      POKER.SetPlayerCookie("playerid", POKER.PLAYER_ID);
      POKER.SetPlayerCookie("playerkey", POKER.PLAYER_KEY);
      document.location.search = "";
    }
    if (!POKER.PLAYER_ID) {
      POKER.PLAYER_ID = POKER.GetPlayerId();
    }

    // Initialize all the renderers
    Signup.Setup();
    TableRenderer.Setup();
  },
  SetPlayerCookie: function(name, val) {
    // We set this in a cookie.
    var d = new Date();
    d.setTime(d.getTime() + (365*24*60*60*1000));
    var expires = "expires="+ d.toUTCString();
    document.cookie =  name + "=" + val + ";" + expires + ";path=/table/" + POKER.NAME;
  },
  UpdateSettings: function(settings) {
    TableRenderer.UpdateSettings(settings);
  },
  GetPlayerId: function() {
    if (!POKER.PLAYER_ID) {
      var m = document.cookie.match(/playerid=(\w+)/)
      if (m) {
        POKER.PLAYER_ID = m[1];
      }
    }
    return POKER.PLAYER_ID;
  },
  UpdateEvents: function(evts) {
  },
  UpdateFullState: function(state) {
    if (state.State == "NOGAME") {
      // Listing of players currently registered, and ability to register.
      Signup.Start(state);
    } else if (state.State == "CASHGAME") {
      // Tables, can join/register.
    } else if (state.State == "SITNGO") {
      // Tables, no joining/registering.
    }
  },
  Poll: function(eventid) {
    var params = {
      Last: eventid,
      Name: POKER.NAME,
    };
    params.PlayerId = POKER.GetPlayerId();
    console.log("Polling");
    $.ajax({
      url: '/GetState',
      type: 'POST',
      dataType: 'json',
      data: JSON.stringify(params),
      success: function(result) {
        console.log(result);
        if (result == "false") {
          setTimeout(function() {
            POKER.Poll(eventid);
          }, 500);
        } else {
          console.log("unblank");
          POKER.UpdateFullState(result);
        }
      }
    });
  },
  SendCommand: function(command, args) {
    var params = {
      Name: POKER.NAME,
      PlayerId: POKER.GetPlayerId(),
      Command: command,
      Args: args,
    };
    $.ajax({
      url: '/Poker',
      type: 'POST',
      dataType: 'json',
      data: JSON.stringify(params),
      success: function(result) {
        console.log(result);
      }
    });
  },
};

$(document).ready(function() {
  POKER.Setup();
  POKER.Poll(0);
});
