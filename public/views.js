// VIEWS contains the renderers
var VIEWS = {};

function View(attrs) {
  for (var name in attrs) {
    this[name] = attrs[name];
  }
  this._subviews = [];
};

View.MakeTemplate = function(elid) {
  var tpl = _.template($(elid).html());
  return function(args) {
    return tpl(args).replace(/\s+</g,'<').replace(/>\s+/g,'>');
  }
};

View.prototype.init = function() {
  this.Cache = {};
  if (this.Templates) {
    this.T = {};
    for (var tplname in this.Templates) {
      this.T[tplname] = View.MakeTemplate(this.Templates[tplname]);
    }
  }
};

View.prototype.AddSubview = function(elementid, view) {
  if (view.init) { view.init(); }
  var el = $(elementid);
  view.Start(el);
  this._subviews.push({elid: elementid, view: view});
}

View.prototype.RemoveSubview = function(elementid) {
  var sv = _.find(this._subviews, function(i) { return i.elid == elementid; });
  this._subviews = _.filter(this._subviews, function(i) { return i.elid != elementid; });
}

View.prototype._handleEvent = function(name, evt) {
  // name == "Click", "Submit", etc. UpperCamelCase.
  var targ = evt.target;
  var handlers = this["On" + name];
  if (handlers) {
    while (targ != null && targ != undefined) {
      if (handlers[targ.id]) {
        evt.target = targ;
        if (handlers[targ.id].call(this, evt) != true) {
          evt.preventDefault();
          evt.stopPropagation();
          return true
        }
      }
      targ = targ.parentElement;
    }
  }
  for (var i in this._subviews) {
    var sv = this._subviews[i].view;
    if (sv._handleEvent) {
      if (sv._handleEvent(name, evt)) { return true; }
    }
  }
};

View.prototype.LocalEvent = function(name, args) {
  this._expectedEvent = [name, args];
  if (this.OnEvent[name]) {
    console.log("Local event triggered");
    this.OnEvent[name].apply(this, args);
  } else {
    console.log("No Events[" + name + "]!");
  }
}

View.prototype.TriggerEvent = function(name, args) {
  if (this._expectedEvent) {
    if (name == this._expectedEvent[0] && _.isEqual(args, this._expectedEvent[1])) {
      this._expectedEvent = undefined;
      return;
    } else {
      console.log("Oops, Remote event doesn't equal Local");
      console.log(this._expectedEvent);
      console.log([name, args]);
      this._expectedEvent = undefined;
    }
  }
  if (this.OnEvent[name]) {
    this.OnEvent[name].apply(this, args);
  } else {
    console.log("No Events[" + name + "]!");
  }
};

View.prototype.ValueDiffers = function(name, val) {
  var ret = this.Differs(this.Cache[name], val);
  this.Cache[name] = val;
  return ret;
};

View.prototype.Differs = function(alpha, beta) {
  if (typeof alpha == "object") {
    for (var a in alpha) {
      if (this.Differs(alpha[a], beta[a])) return true
    }
    for (var b in beta) {
      if (this.Differs(alpha[b], beta[b])) return true
    }
    return false
  } else {
    return alpha != beta
  }
};
