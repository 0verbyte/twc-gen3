var UTILS = {
  renderBool: (b) => {
    return b ? "true" : "false";
  },

  doneTyping: () => {
    let timer = null;
    return function (fn, ms) {
      clearTimeout(timer);
      timer = setTimeout(fn, ms);
    };
  },
};

export default UTILS;
