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
  HandleEvents: function() {
    function bind(ename, name) {
      $(window).on(ename, function(evt) {
        RollPoker.ResetActivity();
        if (RollPoker.Handler._handleEvent) {
          RollPoker.Handler._handleEvent(name, evt);
        }
      });
    }
    var eventmap = {
      "click touchstart": "Click",
      "submit": "Submit",
      "change": "Change",
      "mousemove": "MouseMove",
      "mousedown": "MouseDown",
      "mouseup": "MouseUp",
      "keydown": "KeyDown",
      "keyup": "KeyUp",
    };
    for (var ename in eventmap) {
      bind(ename, eventmap[ename]);
    }
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
  Update: function(doc) {
    Player.state = undefined;
    if (doc.Players) {
      Player.info = doc.Players[Player.uid];
    }
    Game.data = doc;
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
      RollPoker.ProcessLogs(logs, false);
      // Then start a tail.
      Game.watchers.logs = logref.orderBy("Timestamp", "desc").limit(1).onSnapshot(function(logs) {
        RollPoker.ProcessLogs(logs, true);
      });
    });
  },
  LATEST_SEEN: 0,
  UpdateLog: function(log) {
    if (this.Handler.OnLog) {
      RollPoker.Handler.OnLog(log.Message);
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
      if (RollPoker.LATEST_SEEN < litems.Timestamp) {
        if (litems.Logs) {
          RollPoker.LATEST_SEEN = litems.Timestamp;
          for (var j = 0; j < litems.Logs.length; j++) {
            var litem = litems.Logs[j];
            if (litem.Message && litem.Message != "") {
              RollPoker.UpdateLog(litem);
            } else if (doevents) {
              if (RollPoker.Handler._gameEvent) {
                RollPoker.Handler._gameEvent(litem.EventName, litem.Args);
              }
            }
          }
        }
      }
    }
  },
};

$(document).ready(function() {
  RollPoker.SetupFirebase();
  RollPoker.Auth(function() {
    RollPoker.Setup();
  });
});
