// KubeSphere extension entry. [[ .Name ]] is rendered at ksbuilder create time.
// iframe loads index.html via ReverseProxy (/proxy/<name>/*) - same origin as ks-console, no CORS.
System.register(["react", "@kubed/components"], (function (exports) {
  "use strict";
  var React, Loading, EXT_NAME;
  return {
    setters: [
      function (m) { React = m.default; },
      function (m) { Loading = m.Loading; }
    ],
    execute: function () {
      EXT_NAME = "[[ .Name ]]";
      function IframeApp() {
        var _a = React.useState(true), loading = _a[0], setLoading = _a[1];
        return React.createElement(React.Fragment, null,
          loading && React.createElement(Loading, { className: "page-loading" }),
          React.createElement("iframe", {
            src: "/proxy/" + EXT_NAME + "/index.html",
            width: "100%",
            height: "100%",
            frameBorder: "0",
            style: { height: "calc(100vh - 68px)", display: loading ? "none" : "block" },
            onLoad: function () { setLoading(false); }
          })
        );
      }
      exports("default", {
        routes: [{ path: "/" + EXT_NAME, element: React.createElement(IframeApp, null) }],
        menus: [{ parent: "global", name: EXT_NAME, link: "/" + EXT_NAME, title: EXT_NAME, icon: "cluster", order: 0, skipAuth: true }],
        locales: {}
      });
    }
  };
}));
