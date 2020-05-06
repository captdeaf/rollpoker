var VidChat = {
  init: function(uid) {
    // Ugh, different browsers, different approaches.
    VidChat.InitStream();
    VidChat.PeerId = "rpkr" + uid;
    var opts = {
      'iceServers': [{ 'urls': 'stun:stun.l.google.com:19302' }],
    };
    VidChat.Peer = new Peer(VidChat.PeerId);
    VidChat.Peer.on('open', function(pid) {
      console.log("on open");
      console.log(arguments);
      VidChat.Ready();
    });
    VidChat.Connections = {};
    VidChat.Vids = {};
  },
  Update: function() {
    for (var peerid in VidChat.Vids) {
      VidChat.MoveVideo(VidChat.Vids[peerid], peerid);
    }
  },
  InitStream: function() {
    var opts = {audio: true, video: {width: 120, height: 120}};
    var callback = function(stream) {
      VidChat.UserStream = stream;
      VidChat.ConnectStream(VidChat.UserStream, "rpkr" + Player.uid);
      VidChat.Listen();
    };
    var errback = function() {
      alert("Could not initialize video");
    };
    if (navigator.getUserMedia) {
      navigator.getUserMedia(opts, callback, errback);
      return;
    }
    if (navigator.webkitGetUserMedia) {
      navigator.webkitGetUserMedia(opts, callback, errback);
      return;
    }
    if (navigator.mediaDevices && navigator.mediaDevices.getUserMedia) {
      navigator.mediaDevices.getUserMedia(opts).then(callback).catch(errback);
      return;
    }
    if (navigator.mozGetUserMedia) {
      navigator.mozGetUserMedia(opts, callback, errback);
      return;
    }
    alert("Cannot initialize your video");
  },
  IsReady: false,
  Connect: function(remoteid) {
    if (!VidChat.IsReady || !VidChat.UserStream) { return; }
    if (VidChat.Connections[remoteid]) return;
    // To prevent everybody calling each other, we ensure only
    // to call alphabetically lower IDs
    if (VidChat.PeerId < remoteid) { return; }
    console.log(VidChat.PeerId + " calling " + remoteid);
    var call = VidChat.Peer.call(remoteid, VidChat.UserStream);
    if (call) {
      VidChat.Connections[call.peer] = call;
      VidChat.AddCallHandlers(call);
    }
  },
  Ready: function() {
    VidChat.IsReady = true;
    VidChat.Listen();
  },
  Listen: function() {
    if (!VidChat.IsReady || !VidChat.UserStream) {
      return;
    }
    VidChat.Peer.on("call", function(call) {
      VidChat.Connections[call.peer] = call;
      console.log("on call");
      console.log(arguments);
      VidChat.AddCallHandlers(call);
      call.answer(VidChat.UserStream);
      console.log("answered w/ userstream");
    });
  },
  MoveVideo: function(video, peerid) {
    var uid = peerid;
    var m = peerid.match(/^rpkr(.*)/);
    if (m) { uid = m[1]; } else { return; }
    var placeholder = $("#vid-" + uid);
    // TODO: Enable videos in signup, too, as well as spectators.
    if (placeholder.length > 0) {
      var off = placeholder.offset();
      var bbox = placeholder[0].getBoundingClientRect();
      var vidstyle = {
        "position": "absolute",
        "left": bbox.left-1, // Border
        "top": bbox.top,
        "width": bbox.width,
        "height": bbox.height,
      };
      console.log(vidstyle);
      console.log(bbox);
      VidChat.Vids[peerid].css(vidstyle);
    }
  },
  ConnectStream: function(stream, peerid) {
    if (VidChat.Vids[peerid]) { return; }
    VidChat.Vids[peerid] = $('<video id="' + peerid + '" class="playercamera">');
    var vid = VidChat.Vids[peerid][0];
    if ('srcObject' in vid) {
      vid.srcObject = stream;
      vid.play();
    } else {
      vid.src = window.URL.createObjectURL(stream);
      vid.play();
    }
    VidChat.MoveVideo(vid, peerid);
    $('body').append(VidChat.Vids[peerid]);
  },
  AddCallHandlers: function(call) {
    call.on("stream", function(stream) {
      VidChat.ConnectStream(stream, call.peer);
    });
    call.on("close", function() {
      VidChat.Connections[call.peer] = undefined;
      VidChat.Vids[call.peer].remove();
      VidChat.Vids[call.peer] = undefined;
    });
  },
};
