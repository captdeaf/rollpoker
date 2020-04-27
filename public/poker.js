
var Poker = {
  SetupFirebase: function() {
    // Your web app's Firebase configuration
    var firebaseConfig = {
      apiKey: "AIzaSyAsTsJ7UjBQu8CMADJP-JFysn6ON8Hm77M",
      authDomain: "rollpoker.firebaseapp.com",
      databaseURL: "https://rollpoker.firebaseio.com",
      projectId: "rollpoker",
      storageBucket: "rollpoker.appspot.com",
      messagingSenderId: "413322307823",
      appId: "1:413322307823:web:2d12f3485f45d55b12d31a"
    };
    // Initialize Firebase
    firebase.initializeApp(firebaseConfig);
  },
  Setup: function() {
    // First make sure we have our game name.
    var m = document.location.pathname.match(/table\/(\w+)$/);
    if (m) {
      Poker.NAME = m[1];
    } else {
      return;
    }
    var m = document.location.search.match(/\?id=([\w-]+)\&key=([\w-]+)$/);
    if (m) {
      Poker.PLAYER_ID = m[1];
      Poker.PLAYER_KEY = m[2];
      Poker.SetPlayerCookie("playerid", Poker.PLAYER_ID);
      Poker.SetPlayerCookie("playerkey", Poker.PLAYER_KEY);
      // document.location.search = "";
    }
    if (!Poker.PLAYER_ID) {
      Poker.PLAYER_ID = Poker.GetPlayerId();
      Poker.PLAYER_KEY = Poker.GetPlayerKey();
    }

    // Initialize all the renderers
    Signup.Setup();
    Table.Setup();
  },
  SetPlayerCookie: function(name, val) {
    // We set this in a cookie.
    var d = new Date();
    d.setTime(d.getTime() + (365*24*60*60*1000));
    var expires = "expires="+ d.toUTCString();
    var newcookie = name + "=" + val + ";" + expires + ";path=" + document.location.pathname;
    document.cookie = newcookie;
    console.log("Set:", newcookie)
  },
  UpdateSettings: function(settings) {
    Table.UpdateSettings(settings);
  },
  GetPlayerId: function() {
    var m = document.cookie.match(/playerid=(\w+)/)
    if (m) {
      return m[1];
    }
  },
  GetPlayerKey: function() {
    var m = document.cookie.match(/playerkey=(\w+)/)
    if (m) {
      return m[1];
    }
  },
  LAST_STATE: "NOSTATE",
  DATA: {},
  Update: function(doc) {
    Poker.PLAYER = undefined;
    _.each(doc.Players, function(p) {
      if (p.PlayerId == Poker.PLAYER_ID) {
        Poker.PLAYER = p;
      }
    });
    var isNew = false;
    if (Poker.LAST_STATE != doc.State) {
      isNew = true;
      Poker.LAST_STATE = doc.State;
    }
    Poker.DATA = doc;
    var handler = Table;
    if (doc.State == "NOGAME") {
      // Listing of players currently registered, and ability to register.
      handler = Signup;
    }
    if (isNew) {
      Poker.LogCallback = undefined;
      handler.Start(doc);
    }
    handler.Update(doc);
  },
  SendCommand: function(command, args) {
    var params = {
      Name: Poker.NAME,
      PlayerId: Poker.PLAYER_ID,
      PlayerKey: Poker.PLAYER_KEY,
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
  Monitor: function() {
    // Start monitoring the state document.
    if (Poker.NAME && Poker.NAME != "") {
      var db = firebase.firestore();
      Poker.DOCUMENT = db.doc("/public/" + Poker.NAME);
      Poker.DOCUMENT.onSnapshot(function(doc) {
        Poker.Update(doc.data());
      });

      Poker.LOGS = db.collection("/public/" + Poker.NAME + "/log")
      Poker.LOGS.orderBy("Timestamp", "desc").limit(30).get().then(function(logs) {
        Poker.ProcessLogs(logs, false);
        // Then start a tail.
        Poker.LOGS.orderBy("Timestamp", "desc").limit(1).onSnapshot(function(logs) {
          Poker.ProcessLogs(logs, true);
        });
      });
    }
  },
  LATEST_SEEN: 0,
  LogCallback: undefined,
  UpdateLog: function(log) {
    if (Poker.LogCallback) {
      Poker.LogCallback(log.Message);
    }
  },
  ProcessLogs: function(logs, doevents) {
    // We get them in an ordered descent. Reverse 'em.
    var rev = [];
    logs.forEach(function(log) {
      rev.push(log.data());
    });
    for (var i = rev.length - 1; i >= 0; i--) {
      var litems = rev[i];
      if (Poker.LATEST_SEEN < litems.Timestamp) {
        if (!litems.Logs) {
          console.log("Unknown log items", litems);
        } else {
          Poker.LATEST_SEEN = litems.Timestamp;
          for (var j = 0; j < litems.Logs.length; j++) {
            var litem = litems.Logs[j];
            if (litem.Message && litem.Message != "") {
              Poker.UpdateLog(litem);
            } else if (doevents) {
              var evt = Events[litem.EventName];
              if (evt) {
                evt.apply(evt, litem.Args);
              } else {
                console.log("No Events[" + litem.EventName + "]!");
              }
            }
          }
        }
      }
    }
  },
  GetMyTable: function() {
    // TODO: MTTs
    return "table0";
  },
  GetPlayerLocation: function(playerid) {
    return $("#" + playerid).offset();
  },
  GetPlayerSeat: function(playerid) {
    // returns undefined if not seated at same table as we are watching
    var mytable = Poker.DATA.Tables[Poker.GetMyTable()]
    for (var seatname in mytable.Seats) {
      var pid = mytable.Seats[seatname];
      if (pid == playerid) {
        return seatname;
      }
    };
    // Not at this table.
    return undefined;
  },
};

$(document).ready(function() {
  Poker.SetupFirebase();
  Poker.Setup();
  Poker.Monitor();
});
