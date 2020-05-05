var Game = {
  name: "", // Game.name is table/{name}/...
  data: {}, // The full game structure.
  watchers: {}, // Firestore watchers / onSnapshots.
};

var Player = {
  uid: "",   // UID, set from firebase auth
  state: "", // Player[UID].State from game structure
  info: null,  // All info from player[uid].state
  pdata: {}, // From game/data/<uid>
};

var Presences = {};

var RollPoker = {
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
  Auth: function(cb) {
    firebase.auth().onAuthStateChanged(function(user) {
      if (user) {
        console.log("Registered as " + user.displayName);
        Player.uid = user.uid;
        Player.authuser = user;
        cb();
      } else {
        console.log("Not registered");
        RollPoker.Update({
          RoomState: "Register",
          Players: [],
        });
      }
    });
  },
  Scale: 1.0,
  Scales: "scale(1.0)",
  Resize: function() {
    // Portrait and Landscape.
    var win = $(window);
    var rWidth = 600;
    if (win.width() < win.height()) {
      // Portrait, we make width fit.
      var wantedWidth = 590;
      var maxWidth = $(window).width();
      RollPoker.Scale = (Math.floor((maxWidth * 100)/ wantedWidth)) / 100;
    } else {
      // Landscape, we need to fit width to 1/2 the width
      rWidth = 1200;
      var wantedWidth = 590;
      var maxWidth = $(window).width() / 2;
      RollPoker.Scale = (Math.floor((maxWidth * 100)/ wantedWidth)) / 100;
    }
    if (RollPoker.Scale > 1.0) RollPoker.Scale = 1.0;
    RollPoker.Scales = "scale(" + RollPoker.Scale + ")";
    $("#sizer").css({
      "-webkit-transform": RollPoker.Scales,
      "-moz-transform": RollPoker.Scales,
      "-ms-transform": RollPoker.Scales,
      "-o-transform": RollPoker.Scales,
      "transform": RollPoker.Scales,
      "width": rWidth,
    });
  },
  Setup: function() {
    // First make sure we have our game name.
    var m = document.location.pathname.match(/table\/(\w+)$/);
    if (m) {
      Game.name = m[1];
    } else {
      return;
    }

    // Start monitoring.
    RollPoker.Monitor();
    RollPoker.HandleEvents();
    RollPoker.ResetActivity();
    RollPoker.UpdateActivity();
    RollPoker.UpdatePresences();
  },
  Timestamp: function() {
    return Math.floor(new Date().getTime() / 1000);
  },
  LAST_ACTIVITY: 0,
  ResetActivity: function() {
    RollPoker.LAST_ACTIVITY = RollPoker.Timestamp();
  },
  UpdateActivity: function() {
    var actref = RollPoker.DB.doc("/games/" + Game.name + "/act/" + Player.uid);
    actref.set({
      timestamp: RollPoker.Timestamp(),
      activity: RollPoker.LAST_ACTIVITY,
    });
  },
  UpdatePresences: function() {
    var actref = RollPoker.DB.collection("/games/" + Game.name + "/act/");
    actref.get().then(function(docs) {
      var changed = false;
      docs.forEach(function(doc) {
        // Less than 3 minutes idle = awake
        // Less than 10 idle = idle
        // Less than 30 idle = asleep
        // Else offline
        var pid = doc.id; // Player UID
        var activity = doc.data();
        var now = RollPoker.Timestamp();
        var idle_since = now - activity.activity;
        var presence = "Offline";
        // If activity.timestamp is more than 1 minute old,
        // player is likely offline.
        if (activity.timestamp > (now - 60)) {
          if (idle_since < 180) {
            presence = "Active";
          } else if (idle_since < 600) {
            presence = "Idle";
          } else if (idle_since < 1800) {
            presence = "Asleep";
          }
        }
        if (Presences[pid] != presence) {
          changed = true;
          Presences[pid] = presence;
        }
      });
      if (changed) {
        RollPoker.Update(Game.data);
      }
    });
  },
  MOUSE_HANDLER: undefined,
  SetMouseHandler: function(obj) {
    RollPoker.MOUSE_HANDLER = obj;
  },
  HandleEvents: function() {
    // Simple events:
    function bind(ename, name) {
      $(window).on(ename, function(evt) {
        RollPoker.ResetActivity();
        if (RollPoker.Handler && RollPoker.Handler._handleEvent) {
          RollPoker.Handler._handleEvent(name, evt);
        }
      });
    }
    var eventmap = {
      "tap click": "Click",
      "submit": "Submit",
      "change": "Change",
      "keydown": "KeyDown",
      "keyup": "KeyUp",
    };
    for (var ename in eventmap) {
      bind(ename, eventmap[ename]);
    }
    // Mouse and Touch events, we roll our own "click" because of
    // the need for delayed activation w/ the Fold, Call and Bet buttons.
    // Every half second we tick the Handler, in case of
    // timed events (such as countdown for tournament blinds,
    // and idle players not responding.)
    RollPoker.TIMER = setInterval(function() {
      if (RollPoker.Handler && RollPoker.Handler.OnSecond) {
        RollPoker.Handler.OnSecond();
      }
    }, 500);
    // Every 15 seconds we update our "Presence" document.
    RollPoker.ACTIVITY_TIMER = setInterval(function() {
      // If we are active within past 30 minutes, we update.
      // Otherwise we're equivalent to offline.
      var now = RollPoker.Timestamp();
      if ((now - RollPoker.LAST_ACTIVITY) < (30*60)) {
        RollPoker.UpdateActivity();
      }
    }, 15000);
    // And if we're not idle, every 30 seconds we update
    // all player presence from documents.
    RollPoker.PRESENCE_TIMER = setInterval(function() {
      var now = RollPoker.Timestamp();
      if ((now - RollPoker.LAST_ACTIVITY) < (30*60)) {
        RollPoker.UpdatePresences();
      }
    }, 30000)
  },
  IsHost: false,
  Update: function(doc) {
    Player.state = undefined;
    if (doc.Players) {
      Player.info = doc.Players[Player.uid];
    }
    Game.data = doc;
    RollPoker.IsHost = false;
    if (Game.data && Game.data.Hosts && Game.data.Hosts[Player.uid]) {
      RollPoker.IsHost = Game.data.Hosts[Player.uid];
    }
    if (Game.LAST_STATE != doc.RoomState) {
      Game.LAST_STATE = doc.RoomState;
      RollPoker.Handler = VIEWS[doc.RoomState];
      if (RollPoker.Handler) {
        RollPoker.Handler.init();
        RollPoker.Handler.Start();
      }
    }
    if (RollPoker.Handler) {
      RollPoker.Handler.Update(doc);
    }
    if (doc.LastLogs) {
      if (RollPoker.RunLogs) {
        RollPoker.ProcessLogs([doc.LastLogs], true);
      } else {
        RollPoker.BackLog.push(doc.LastLogs);
      }
    }
  },
  SendCommand: function(command, args, onsucc) {
    var params = {
      Name: Game.name,
      Command: command,
      Args: args,
    };
    var headers = {};
    Player.authuser.getIdToken().then(function(token) {
      $.ajax({
        url: '/Poker',
        type: 'POST',
        dataType: 'text',
        headers: {
          Authorization: "Bearer " + token,
        },
        data: JSON.stringify(params),
        success: function(result) {
          if (onsucc) {
            onsucc();
          }
        },
        error: function(result) {
          console.log("err", result.responseText);
          alert(result.responseText);
        },
      });
    });
  },
  Monitor: function() {
    // Start monitoring the state documents.
    RollPoker.DB = firebase.firestore();
    var docref = RollPoker.DB.doc("/games/" + Game.name);
    Game.watchers.data = docref.onSnapshot(function(doc) {
      RollPoker.Update(doc.data());
    }, function(error) {
      RollPoker.Update({
        RoomState: "Signup",
        Players: [],
      });
    });
    var dataref = RollPoker.DB.doc("/games/" + Game.name + "/data/" + Player.uid);
    Game.watchers.data = dataref.onSnapshot(function(doc) {
      // In theory, pdata is always updated immediately before a
      // document update.
      Player.pdata = doc.data();
      RollPoker.Update(Game.data);
    });

    var logref = RollPoker.DB.collection("/games/" + Game.name + "/log")
    logref.orderBy("Timestamp", "desc").limit(30).get().then(function(logs) {
      // We get them in an ordered descent. Reverse 'em and convert to data.
      var revlogs = [];
      logs.forEach(function(log) {
        revlogs.push(log.data());
      });
      RollPoker.ProcessLogs(revlogs, false);
      RollPoker.RunLogs = true;
      RollPoker.ProcessLogs(RollPoker.BackLog, true);
    });
  },
  LATEST_SEEN: 0,
  UpdateLog: function(log) {
    if (this.Handler.OnLog) {
      RollPoker.Handler.OnLog(log.Message);
    }
  },
  RunLogs: false,
  BackLog: [],
  ProcessLogs: function(logs, doevents) {
    for (var i = logs.length - 1; i >= 0; i--) {
      var litems = logs[i];
      if (RollPoker.LATEST_SEEN < litems.Timestamp) {
        if (litems.Logs) {
          RollPoker.LATEST_SEEN = litems.Timestamp;
          for (var j = 0; j < litems.Logs.length; j++) {
            var litem = litems.Logs[j];
            if (litem.Message && litem.Message != "") {
              RollPoker.UpdateLog(litem);
            } else if (doevents) {
              RollPoker.TriggerEvent(litem.EventName, litem.Args);
            }
          }
        }
      }
    }
  },
  TriggerEvent: function(evt, args) {
    if (RollPoker.Handler.TriggerEvent) {
      RollPoker.Handler.TriggerEvent(evt, args);
    }
  },
  LocalEvent: function(evt, args) {
    if (RollPoker.Handler.LocalEvent) {
      RollPoker.Handler.LocalEvent(evt, args);
    }
  },
};

$(document).ready(function() {
  RollPoker.SetupFirebase();
  RollPoker.Auth(function() {
    RollPoker.Setup();
  });
  RollPoker.Resize();
  $(window).resize(RollPoker.Resize);
});
