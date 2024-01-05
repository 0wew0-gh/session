export default class Session {
  api = "http://192.168.1.46:36040";
  constructor() {
    // 留言板
    const leaveMsg = document.getElementById(
      "leaveMsg"
    ) as HTMLDivElement | null;
    const lmTopBtn = document.getElementById(
      "lm-top-btn"
    ) as HTMLDivElement | null;
    const lmTopBtnR = document.getElementById(
      "lm-top-btn-r"
    ) as HTMLDivElement | null; // 留言板打开和关闭
    if (lmTopBtn != null && leaveMsg != null) {
      lmTopBtn.addEventListener("click", function () {
        if (leaveMsg.classList.contains("open")) {
          leaveMsg.classList.remove("open");
          leaveMsg.classList.add("close");
          if (lmTopBtnR != null) lmTopBtnR.innerHTML = "^";
        } else {
          leaveMsg.classList.remove("close");
          leaveMsg.classList.add("open");
          if (lmTopBtnR != null) lmTopBtnR.innerHTML = "v";
        }
      });
    }

    const name = document.getElementById(
      "lm-form-name"
    ) as HTMLInputElement | null;
    const nameWarning = document.getElementById(
      "lm-name-warning"
    ) as HTMLParagraphElement | null;
    const mail = document.getElementById(
      "lm-form-mail"
    ) as HTMLInputElement | null;
    const mailWarning = document.getElementById(
      "lm-mail-warning"
    ) as HTMLParagraphElement | null;
    const phone = document.getElementById(
      "lm-form-phone"
    ) as HTMLInputElement | null;
    const phoneWarning = document.getElementById(
      "lm-phone-warning"
    ) as HTMLParagraphElement | null;
    const msg = document.getElementById(
      "lm-form-msg"
    ) as HTMLTextAreaElement | null;
    const msgWarning = document.getElementById(
      "lm-msg-warning"
    ) as HTMLParagraphElement | null;
    const msgWrapper = document.getElementById(
      "lm-form-msg-wrapper"
    ) as HTMLDivElement | null;

    name?.addEventListener("input", () => {
      if (name.value.length > 255) {
        if (nameWarning != null) {
          nameWarning.innerText = "Name too long";
        }
        this.objectSwitch(
          name,
          "lm-input-border-color-err",
          "lm-input-border-color"
        );
      } else {
        if (nameWarning != null) {
          nameWarning.innerText = "";
        }
        this.objectSwitch(
          name,
          "lm-input-border-color",
          "lm-input-border-color-err"
        );
      }
    });
    mail?.addEventListener("input", () => {
      if (mailWarning != null)
        mailWarning.innerText = mail.value.length + "/150";
      if (mail.value.length > 150) {
        if (mailWarning != null) {
          mailWarning.innerText = "Email too long";
          mailWarning.classList.add("warning-err");
        }
        this.objectSwitch(
          mail,
          "lm-input-border-color-err",
          "lm-input-border-color"
        );
      } else {
        if (mailWarning != null) mailWarning.classList.remove("warning-err");
        this.objectSwitch(
          mail,
          "lm-input-border-color",
          "lm-input-border-color-err"
        );
      }
    });
    phone?.addEventListener("input", () => {
      if (phoneWarning != null)
        phoneWarning.innerText = phone.value.length + "/100";
      if (phone.value.length > 100) {
        if (phoneWarning != null) {
          phoneWarning.innerText = "Phone too long";
          phoneWarning.classList.add("warning-err");
        }
        this.objectSwitch(
          phone,
          "lm-input-border-color-err",
          "lm-input-border-color"
        );
      } else {
        if (phoneWarning != null) phoneWarning.classList.remove("warning-err");
        this.objectSwitch(
          phone,
          "lm-input-border-color",
          "lm-input-border-color-err"
        );
      }
    });
    msg?.addEventListener("input", () => {
      if (msgWarning != null) msgWarning.innerText = msg.value.length + "/500";
      if (msg.value.length > 500) {
        if (msgWarning != null) {
          msgWarning.innerText = "Email too long";
          msgWarning.classList.add("warning-err");
        }
        this.objectSwitch(
          msgWrapper,
          "lm-input-border-color-err",
          "lm-input-border-color"
        );
      } else {
        this.objectSwitch(
          msgWrapper,
          "lm-input-border-color",
          "lm-input-border-color-err"
        );
        if (msgWarning != null) msgWarning.classList.remove("warning-err");
      }
    });

    // 留言板发送按钮
    const lmBtnSend = document.getElementById(
      "lm-form-btn-send"
    ) as HTMLButtonElement | null;
    lmBtnSend?.addEventListener("click", () => {
      let isRequired = true;
      if (name != null && name.value == "") {
        this.objectSwitch(
          name,
          "lm-input-border-color-err",
          "lm-input-border-color"
        );
        if (nameWarning != null) nameWarning.innerText = "Name is not empty";
        isRequired = false;
      } else if (name != null && name.value.length > 255) {
        this.objectSwitch(
          name,
          "lm-input-border-color-err",
          "lm-input-border-color"
        );
        isRequired = false;
      } else {
        this.objectSwitch(
          name,
          "lm-input-border-color",
          "lm-input-border-color-err"
        );
        if (nameWarning != null) nameWarning.innerText = "";
      }

      if (mail != null && mail.value == "") {
        this.objectSwitch(
          mail,
          "lm-input-border-color-err",
          "lm-input-border-color"
        );
        if (mailWarning != null) {
          mailWarning.innerText = "Email is not empty";
          mailWarning.classList.add("warning-err");
        }
        isRequired = false;
      } else if (mail != null && mail.value.length > 150) {
        this.objectSwitch(
          mail,
          "lm-input-border-color-err",
          "lm-input-border-color"
        );
        isRequired = false;
      } else {
        this.objectSwitch(
          mail,
          "lm-input-border-color",
          "lm-input-border-color-err"
        );
        if (mailWarning != null) {
          mailWarning.innerText = "";
          mailWarning.classList.remove("warning-err");
        }
      }

      if (phone != null && phone.value.length > 100) {
        this.objectSwitch(
          phone,
          "lm-input-border-color-err",
          "lm-input-border-color"
        );
      } else {
        this.objectSwitch(
          phone,
          "lm-input-border-color",
          "lm-input-border-color-err"
        );
      }

      if (msg != null && msg.value == "") {
        this.objectSwitch(
          msgWrapper,
          "lm-input-border-color-err",
          "lm-input-border-color"
        );
        if (msgWarning != null) {
          msgWarning.innerText = "Message is not empty";
          msgWarning.classList.add("warning-err");
        }
        isRequired = false;
      } else if (msg != null && msg.value.length > 500) {
        this.objectSwitch(
          msgWrapper,
          "lm-input-border-color-err",
          "lm-input-border-color"
        );
        isRequired = false;
      } else {
        this.objectSwitch(
          msgWrapper,
          "lm-input-border-color",
          "lm-input-border-color-err"
        );
        if (msgWarning != null) {
          msgWarning.innerText = "";
          msgWarning.classList.remove("warning-err");
        }
      }
      if (!isRequired) {
        return false;
      }
      if (
        name != null &&
        mail != null &&
        phone != null &&
        msg != null &&
        lmBtnSend != null
      ) {
        this.sendLeaveMsg(
          name.value,
          mail.value,
          phone.value,
          msg.value,
          lmBtnSend
        );
      }
    });
  }

  sendLeaveMsg(
    name: string,
    mail: string,
    phone: string,
    msg: string,
    btnSend: HTMLButtonElement
  ) {
    btnSend.disabled = true;
    const btnSendInnerHTML = btnSend.innerHTML;
    btnSend.innerText = "Sending...";

    const toolTips = document.getElementsByClassName(
      "lm-tooltips"
    ) as HTMLCollection;
    const toolTipsMsg = document.getElementById(
      "lm-tooltips-msg"
    ) as HTMLDivElement | null;

    const formData = new FormData();
    formData.append("n", name);
    formData.append("m", mail);
    if (phone != "") {
      formData.append("p", phone);
    }
    formData.append("msg", msg);
    const currentUrl = window.location.href;
    formData.append("page", currentUrl);
    formData.append("se", "1"); // 开启详细错误信息
    formData.append("l", "2"); // 中文提示

    return fetch(this.api + "/setLeaveMsg/", {
      method: "POST",
      body: formData,
    })
      .then((response) => response.json())
      .then((res) => {
        btnSend.disabled = false;
        btnSend.innerHTML = btnSendInnerHTML;
        if (res.code == 10000) {
          const data = res.msg;
          if (toolTipsMsg != null) {
            toolTipsMsg.innerHTML = data;
          }
        }
        this.objectSwitch(toolTips, "open", "close");
        setTimeout(() => {
          this.objectSwitch(toolTips, "close", "open");
        }, 3000);
      })
      .catch((error) => {
        console.log("error", error);
        btnSend.classList.add("lm-form-background-color-err");
        btnSend.innerHTML = "Send failed";
        this.objectSwitch(toolTips, "open", "close");
        setTimeout(() => {
          btnSend.disabled = false;
          btnSend.classList.remove("lm-form-background-color-err");
          btnSend.innerHTML = btnSendInnerHTML;
          this.objectSwitch(toolTips, "close", "open");
        }, 3000);
        console.error(error);
      });
  }
  objectSwitch(
    selectors: HTMLCollection | HTMLDivElement | string | null,
    open: string,
    close: string
  ) {
    let object: HTMLCollection | HTMLElement | null = null;
    if (typeof selectors == "string") {
      object = document.querySelector(selectors) as HTMLElement | null;
    } else {
      object = selectors;
    }
    if (object == null || object == undefined) {
      return;
    }
    if (object instanceof HTMLCollection) {
      for (let i = 0; i < object.length; i++) {
        object[i].classList.remove(close);
        object[i].classList.add(open);
      }
      return;
    }
    object.classList.remove(close);
    object.classList.add(open);
  }
}
