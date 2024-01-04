interface msgItem{
  id: string;
  sid: string;
  content: string;
  time: number;
  userID: string;
}
interface MessageItem {
  msg: msgItem[];
  user: {
    nick: string;
    mail: string;
  };
}

export default class Session {
  api = "http://192.168.1.46:36040";
  api2 = "http://192.168.1.46:36041";

  openSessionView: HTMLButtonElement = document.getElementById(
    "open-session-view"
  ) as HTMLButtonElement;

  messageList = []; // 消息列表
  userMap: Record<string, string> | null = null; // 用户列表
  selfMail: string | null = null; // 用户邮箱
  selfID: string | null = null; // 用户ID
  sessionID: string | null = null; // 会话ID
  token: string | null = null;
  source: string | null = null; // SSE对象
  isScroll = true; // 是否滚动到最下
  isSending = false; // 是否正在发送消息

  constructor() {
    this.openSessionView.addEventListener("click", () => {
      this.token = localStorage.getItem("token");
      this.selfMail = localStorage.getItem("user");
      this.sessionID = localStorage.getItem("sessionID");
      if (this.token == null) {
        this.getGuestToken()
          .then(() => {
            console.log(">>> getGuestToken Done <<<");
            this.getMessage();
          })
          .catch((error) => {
            console.log(error);
            // TODO: 提示 获取 token 失败
            return;
          });
      } else {
        this.checkToken()?.then((res) => {
          if (res) {
            console.log(">>> checkToken Done <<<");
            this.getMessage();
          } else {
            this.getGuestToken()
              .then(() => {
                console.log(">>> getGuestToken Done <<<");
                this.getMessage();
              })
              .catch((error) => {
                console.log(error);
                // TODO: 提示 获取 token 失败
                return;
              });
          }
        });
      }
    });
  }

  getGuestToken(): Promise<void> {
    this.token = null;
    this.selfMail = null;
    this.sessionID = null;
    localStorage.removeItem("token");
    localStorage.removeItem("user");
    localStorage.removeItem("sessionID");
    const formData: FormData = new FormData();
    formData.append("g", "1");

    return fetch(this.api + "/login/", {
      method: "POST",
      body: formData,
    })
      .then((response) => response.json())
      .then((res) => {
        // console.log(">>>", res);
        if (res.code == 10000) {
          const data = res.data;
          this.token = data.token;
          this.selfMail = data.mail;
          localStorage.setItem("token", this.token ?? "");
          localStorage.setItem("user", this.selfMail ?? "");
          // console.log("> token >", token);
          // console.log("> mail >", selfMail);
        }
      })
      .catch((error) => {
        console.log(error);
      });
  }

  checkToken(): Promise<boolean> | null {
    if (this.token == null) {
      // TODO: 提示 token 为空
      return null;
    }
    const formData = new FormData();
    formData.append("t", this.token);

    return fetch(`${this.api}/checkToken/`, {
      method: "POST",
      body: formData,
    })
      .then((response) => response.json())
      .then((res) => {
        switch (res.code) {
          case 10000:
            return true;
            break;
          default:
            return false;
            break;
        }
      });
  }

  getMessage() {
    if (this.sessionID == null) {
      return;
    }
    const formData = new FormData();
    formData.append("sid", this.sessionID);
    formData.append("o", "time");
    formData.append("osc", "DESC");
    formData.append("lt", "0,200");
    formData.append("l", "2");

    fetch(this.api + "/getMessage/", {
      method: "POST",
      body: formData,
    })
      .then((response) => response.json())
      .then((res) => {
        let data;
        let user;
        let mail;
        switch (res.code) {
          case 10000:
            data = res.data.msg;
            user = res.data.user;
            mail = res.data.mail;
            if (this.messageList.length == 0) {
              this.messageList = data;
            } else {
              this.messageList = this.messageList.concat(data);
            }
            if (this.userMap == null) {
              this.userMap = user;
            } else {
              for (const key in user) {
                if (!Object.prototype.hasOwnProperty.call(this.userMap, key)) {
                  this.userMap[key] = user[key];
                }
              }
            }
            this.handleMessage(messageList);
            setTimeout(function () {
              this.scrollBottom(300);
            }, 350);
            break;
          case 10001:
          case 10002:
          case 10003:
            break;
          case 3900:
            this.token = null;
            localStorage.removeItem("token");
            this.getGuestToken()
              .then(() => {
                console.log(">>> getGuestToken Done <<<");
                this.getMessage();
              })
              .catch((error) => {
                console.log(error);
                // TODO: 提示 获取 token 失败
                return;
              });
        }

        this.setJustifyContent();
      })
      .catch((error) => {
        console.log(error);
      });
  }
  sse() {
    this.sseClose();
    if (this.sessionID == null) {
      return;
    }
    const params = new URLSearchParams({
      sid: this.sessionID,
    });
    const source = new EventSource(`${this.api2}/sse/?${params}`);
    source.onmessage = (event) => {
      // console.log(">>> onmessage <<<");
      // console.log(event.data);
      let ed :MessageItem= {
        msg: [],
        user: {
          nick: "",
          mail: "",
        },
      };
      const newMsgList:msgItem[] = [];
      try {
        ed = JSON.parse(event.data);
      } catch (error) {
        console.log(error);
        return;
      }
      console.log(">> ed <<", ed);
      const msgs = ed.msg;
      const users = ed.user;
      if (this.userMap == null) {
        this.userMap = users;
      } else {
        this.userMap = Object.assign(this.userMap, users);
      }
      for (let i = 0; i < msgs.length; i++) {
        const e = msgs[i];
        let isHave = false;
        // console.log(e);
        const userID = e["userID"];
        if (userID == "") {
          // for (const key in this.userMap) {
          //   if (Object.prototype.hasOwnProperty.call(this.userMap, key)) {
          //     const uMe = this.userMap[key];
          //     if (uMe.mail == e.user.mail) {
          //       isHave = true;
          //       userID = key;
          //       break;
          //     }
          //   }
          // }
          // if (!isHave) {
          //   this.userMap["-1"] = {
          //     nick: e.user.nick,
          //     mail: e.user.mail,
          //   };
          // }
        } else {
          isHave = true;
          if (userMap != null) {
            console.log(">>>", userMap, userID);
            if (!Object.prototype.hasOwnProperty.call(userMap, userID)) {
              userMap[userID] = {
                nick: e.user.nick,
                mail: e.user.mail,
              };
            }
          }
        }
        newMsgList.push({
          id: "-1",
          sid: sessionID,
          content: e.content,
          time: e.time,
          userID: isHave ? userID : "-1",
        });
        // console.log(">>", newMsgList);
      }
      messageList = messageList.concat(newMsgList);
      // console.log("=== sse ===");
      // console.log(newMsgList);
      // console.log(messageList);
      // console.log(userMap);
      handleMessage(newMsgList);
      scrollBottom(300);
    };
  }
}
