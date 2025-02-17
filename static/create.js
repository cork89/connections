"use strict";import{isErrorWithProperty as k}from"./common.js";var N=(a=>(a.NONE="none",a.YELLOW="yellow",a.GREEN="green",a.BLUE="blue",a.PURPLE="purple",a))(N||{});const B={none:"none",yellow:"yellow",green:"green",blue:"blue",purple:"purple"};function O(){return{yellow:{category:"",words:[]},green:{category:"",words:[]},blue:{category:"",words:[]},purple:{category:"",words:[]}}}function x(){return{state:"none",categories:O(),gameId:""}}function D(t,e,r){const o={gameId:t,timestamp:e,categories:r};d.push(o),localStorage.setItem("createdGames",JSON.stringify(d))}const W=document.getElementById("yellow-edit")??(()=>{throw new Error("yellowEdit cannot be null")})();W.addEventListener("click",t=>u(t.target));const q=document.getElementById("green-edit")??(()=>{throw new Error("greenEdit cannot be null")})();q.addEventListener("click",t=>u(t.target));const A=document.getElementById("blue-edit")??(()=>{throw new Error("blueEdit cannot be null")})();A.addEventListener("click",t=>u(t.target));const R=document.getElementById("purple-edit")??(()=>{throw new Error("purpleEdit cannot be null")})();R.addEventListener("click",t=>u(t.target));function u(t){const e=t?.parentElement?.parentElement;if(e&&e.classList[1]==n.state)t.src="/static/edit.svg",L(e);else if(e&&n.state=="none")t.src="/static/x.svg",C(e);else{const r=document.getElementById(`${n.state}-edit`),o=r?.parentElement?.parentElement;r&&o&&e&&(r.src="/static/edit.svg",t.src="/static/x.svg",L(o),C(e))}}var i,g,E,y,d,n,l=[],c=[];function $(){i=document.getElementById("category-input")??(()=>{throw new Error("categoryInput cannot be null")})(),g=document.getElementById("word-input")??(()=>{throw new Error("wordsInput cannot be null")})(),E=document.getElementById("save-categories-button")??(()=>{throw new Error("saveCategoriesButton cannot be null")})(),y=document.getElementById("submit")??(()=>{throw new Error("submitButton cannot be null")})();let t=localStorage.getItem("ctx");t?(n=JSON.parse(t),n.state!=="none"&&(n.state="none",localStorage.setItem("ctx",JSON.stringify(n)))):(n=x(),localStorage.setItem("ctx",JSON.stringify(n))),d=JSON.parse(localStorage.getItem("createdGames")??"[]"),d||localStorage.setItem("createdGames",JSON.stringify(d)),w(),document.getElementById("category-input")?.addEventListener("keyup",function(){T()?i.classList.remove("toolong"):i.classList.contains("toolong")||i.classList.add("toolong")}),document.getElementById("category-input")?.addEventListener("keydown",function(e){const r=this;if(e.key===","||e.key==="Enter"){if(e.preventDefault(),!T())return;const o=r.value.trim();o&&c.length<1&&(M(o,r,!0),r.value="")}else if((e.key==="Backspace"||e.key==="Delete")&&!r.value.trim()){const s=c.pop();H(s?.children[0],!0),e.preventDefault()}}),document.getElementById("word-input")?.addEventListener("keyup",function(){this.value.trim().length>b?g.classList.contains("toolong")||g.classList.add("toolong"):g.classList.remove("toolong")}),document.getElementById("word-input")?.addEventListener("keydown",function(e){const r=this,o=r.value.trim();if(e.key===","||e.key==="Enter"){if(e.preventDefault(),!F())return;o&&l.length<4&&(S(o,r,!0),r.value="")}else if((e.key==="Backspace"||e.key==="Delete")&&!o){const s=l.pop();s&&s.hasChildNodes()&&(I(s.children[0],!1,!0),e.preventDefault())}}),document.addEventListener("keydown",function(e){n.state!="none"||document.getElementById("gameId-input")==document.activeElement||(e.key==="Y"||e.key==="y"?(n.state=="none"&&document.querySelector("div.colors-item.yellow")?.querySelector("img")?.click(),e.preventDefault()):e.key==="G"||e.key==="g"?(n.state=="none"&&document.querySelector("div.colors-item.green")?.querySelector("img")?.click(),e.preventDefault()):e.key==="B"||e.key==="b"?(n.state=="none"&&document.querySelector("div.colors-item.blue")?.querySelector("img")?.click(),e.preventDefault()):(e.key==="P"||e.key==="p")&&(n.state=="none"&&document.querySelector("div.colors-item.purple")?.querySelector("img")?.click(),e.preventDefault()))},!0),h();for(const e in n.categories){const r=document.getElementById(`category-${e}`),o=document.getElementById(`words-${e}`),s=r?.parentElement?.parentElement,a=s?.querySelector("img");r&&s&&a&&r.addEventListener("click",()=>u(a)),o&&s&&a&&o.addEventListener("click",()=>u(a))}}$();function w(){for(const t in n.categories){const e=document.getElementById(`category-${t}`),r=document.getElementById(`words-${t}`);let o="tbd";if(n.categories.hasOwnProperty(t)){const s=t;n.categories[s].category!=""&&(o=`<span class="category category2">${n.categories[s].category}</span>`),e&&(e.innerHTML=o);let a="tbd";const m=n.categories[s].words.length;if(m>0){a="";for(let p=0;p<m;p++)a+=`<span class="category category2">${n.categories[s].words[p]}</span>`}r&&(r.innerHTML=a)}}}function C(t){if(n.state==="none"){let e=t.classList[1];n.state=B[e],t.classList.add("color-selected"),i&&(i.disabled=!1),g.disabled=!1,E.disabled=!1,G(),c.length>0?g.focus():i.focus()}}function L(t){n.state="none",t.classList.remove("color-selected"),i.disabled=!0,g.disabled=!0,E.disabled=!0,v()}function G(){const t=n.state.toString();if(t==="yellow"||t==="green"||t==="blue"||t==="purple"){const e=n.categories[t];e.category!=""&&M(e.category,i);for(let r=0;r<e.words.length;r++)S(e.words[r],g)}}function v(){for(let t=0;t<c.length;t++)H(c[t].children[0]);for(let t=0;t<l.length;t++)I(l[t].children[0],!1);c.splice(0,c.length),l.splice(0,l.length),i.value="",g.value=""}function h(){n.categories.yellow.words.length==4&&n.categories.green.words.length==4&&n.categories.blue.words.length==4&&n.categories.purple.words.length==4?y.disabled=!1:y.disabled=!0}const J=document.getElementById("save-categories-button")??(()=>{throw new Error("saveCategories cannot be null")})();J.addEventListener("click",()=>P());function P(){if(n.state!="none"){let t=c.length>0?c[0].innerText:"";n.categories[n.state].category=t.slice(0,t.indexOf(`
`)),n.categories[n.state].words.splice(0,n.categories[n.state].words.length);for(let o=0;o<l.length;o++){let s=l[o].innerText;n.categories[n.state].words.push(s.slice(0,s.indexOf(`
`)))}const e=document.getElementById("gameId-input");n.gameId=e!=null?e.value.trim().replace(/\s+/g,"-"):"",v();const r=document.querySelector(`div.colors-item.${n.state}`);if(r){const o=r.querySelector("img");o.src="/static/edit.svg",L(r)}localStorage.setItem("ctx",JSON.stringify(n)),w(),h()}}function f(t){if(t==="none")return;const e=n.categories[t];let r=c.length>0?c[0].innerText:"";e.category=r.slice(0,r.indexOf(`
`)),e.words.splice(0,e.words.length);for(let o=0;o<l.length;o++){let s=l[o].innerText;e.words.push(s.slice(0,s.indexOf(`
`)))}localStorage.setItem("ctx",JSON.stringify(n)),w()}const j=document.getElementById("submit-warning-close")??(()=>{throw new Error("submitWarningClose cannot be null")})();j.addEventListener("click",t=>U(t.target));function U(t){const e=t.parentElement;e&&(e.style.visibility="hidden")}function Y(){let t=[];const e=n.categories,r=e.yellow,o=e.green,s=e.blue,a=e.purple;return r.category&&t.push(r.category),o.category&&t.push(o.category),s.category&&t.push(s.category),a.category&&t.push(a.category),t}const _=document.getElementById("submit")??(()=>{throw new Error("submit cannot be null")})();_.addEventListener("click",()=>V());async function V(){const t="";try{const e={yellow:n.categories.yellow,green:n.categories.green,blue:n.categories.blue,purple:n.categories.purple,gameId:n.gameId??""},r=await fetch(t,{headers:{Accept:"application/json","Content-Type":"application/json"},method:"POST",body:JSON.stringify(e)});if(!r.ok)throw new Error(`Response status: ${r.status}`);const o=await r.json();if(o.success===!1){const s=document.getElementById("warning-message");s&&(s.innerText=o.failureReason,s.parentElement&&(s.parentElement.style.visibility="visible")),y.disabled=!0}else if(o.gameId&&o.gameId!=""){localStorage.removeItem("ctx"),D(o.gameId,Date.now(),Y()),window.location.href=`/game/${o.gameId}/`;return}}catch(e){k(e,"message")&&console.error(e.message)}}const b=20;function F(){return!(g.value.length>b)}function S(t,e,r=!1){const o=document.querySelector(".word-container"),s=document.createElement("div");s.className="category";const a=document.createElement("span");a.classList.add("word-remove"),a.innerHTML="&times;",a.addEventListener("click",m=>I(m.target)),s.innerHTML=t,s.appendChild(a),o&&o.insertBefore(s,e),l.push(s),r&&f(n.state)}function I(t,e=!0,r=!1){if(e){let o=l.indexOf(t.parentElement);l.splice(o,1)}t?.parentElement?.remove(),g.focus(),r&&f(n.state)}const K=40;function T(){return!(i.value.length>K)}function M(t,e,r=!1){const o=document.querySelector(".category-container")??(()=>{throw new Error("category-container cannot be null")})(),s=document.createElement("div");s.className="category",s.innerHTML=`${t}<span class="word-remove" onclick="removeCategory(this)">&times;</span>`,o.insertBefore(s,e),c.push(s),r&&f(n.state)}function H(t,e=!1){t.parentElement?.remove(),i.focus(),e&&f(n.state)}
