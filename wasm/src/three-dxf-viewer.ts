// @ts-nocheck
// Derived from @f12o/three-dxf (MIT License). See wasm/node_modules/@f12o/three-dxf/LICENSE.
import * as THREE from "three";
import { BufferGeometry, Color, Float32BufferAttribute, Vector3 } from "three";
import { Text } from "troika-three-text";
import { parseDxfMTextContent } from "@dxfom/mtext";

const troikaDepthDescriptor = Object.getOwnPropertyDescriptor(
	Text.prototype,
	"customDepthMaterial"
);
if (troikaDepthDescriptor && !troikaDepthDescriptor.set) {
	Object.defineProperty(Text.prototype, "customDepthMaterial", {
		get: troikaDepthDescriptor.get,
		set() {},
		configurable: true
	});
}

const troikaDistanceDescriptor = Object.getOwnPropertyDescriptor(
	Text.prototype,
	"customDistanceMaterial"
);
if (troikaDistanceDescriptor && !troikaDistanceDescriptor.set) {
	Object.defineProperty(Text.prototype, "customDistanceMaterial", {
		get: troikaDistanceDescriptor.get,
		set() {},
		configurable: true
	});
}

function decodeUnicodeEscapes(value) {
	return value.replace(/\\U\+([0-9A-Fa-f]{4,6})/g, (match, hex) => {
		const codePoint = Number.parseInt(hex, 16);
		if (!Number.isFinite(codePoint) || codePoint < 0 || codePoint > 0x10ffff) {
			return match;
		}
		return String.fromCodePoint(codePoint);
	});
}
function OrbitControls(t, n) {
	this.object = t, this.domElement = n === void 0 ? document : n, this.enabled = !0, this.target = new THREE.Vector3(), this.center = this.target, this.noZoom = !1, this.zoomSpeed = 1, this.minDistance = 0, this.maxDistance = Infinity, this.noRotate = !1, this.rotateSpeed = 1, this.noPan = !1, this.keyPanSpeed = 7, this.autoRotate = !1, this.autoRotateSpeed = 2, this.minPolarAngle = 0, this.maxPolarAngle = Math.PI, this.noKeys = !1, this.keys = {
		LEFT: 37,
		UP: 38,
		RIGHT: 39,
		BOTTOM: 40
	};
	var r = this, i = new THREE.Vector2(), a = new THREE.Vector2(), o = new THREE.Vector2(), s = new THREE.Vector2(), c = new THREE.Vector2(), l = new THREE.Vector2(), u = new THREE.Vector3(), d = new THREE.Vector3(), f = new THREE.Vector2(), p = new THREE.Vector2(), m = new THREE.Vector2(), h = 0, g = 0, _ = 1, v = new THREE.Vector3();
	new THREE.Vector3();
	var y = {
		NONE: -1,
		ROTATE: 0,
		DOLLY: 1,
		PAN: 2,
		TOUCH_ROTATE: 3,
		TOUCH_DOLLY: 4,
		TOUCH_PAN: 5
	}, b = y.NONE;
	this.target0 = this.target.clone(), this.position0 = this.object.position.clone();
	var x = { type: "change" }, S = { type: "start" }, C = { type: "end" };
	this.rotateLeft = function(e) {
		e === void 0 && (e = w()), g -= e;
	}, this.rotateUp = function(e) {
		e === void 0 && (e = w()), h -= e;
	}, this.panLeft = function(e) {
		var t = this.object.matrix.elements;
		u.set(t[0], t[1], t[2]), u.multiplyScalar(-e), v.add(u);
	}, this.panUp = function(e) {
		var t = this.object.matrix.elements;
		u.set(t[4], t[5], t[6]), u.multiplyScalar(e), v.add(u);
	}, this.pan = function(e, t) {
		var n = r.domElement === document ? r.domElement.body : r.domElement;
		if (r.object.fov !== void 0) {
			var i = r.object.position.clone().sub(r.target).length();
			i *= Math.tan(r.object.fov / 2 * Math.PI / 180), r.panLeft(2 * e * i / n.clientHeight), r.panUp(2 * t * i / n.clientHeight);
		} else r.object.top === void 0 ? console.warn("WARNING: OrbitControls.js encountered an unknown camera type - pan disabled.") : (r.panLeft(e * (r.object.right - r.object.left) / n.clientWidth), r.panUp(t * (r.object.top - r.object.bottom) / n.clientHeight));
	}, this.dollyIn = function(e) {
		e === void 0 && (e = T()), _ /= e;
	}, this.dollyOut = function(e) {
		e === void 0 && (e = T()), _ *= e;
	}, this.update = function() {
		r.object.top !== void 0 && (this.object.top = _ * this.object.top, this.object.bottom = _ * this.object.bottom, this.object.left = _ * this.object.left, this.object.right = _ * this.object.right, this.object.updateProjectionMatrix());
		var e = this.object.position;
		d.copy(e).sub(this.target), this.target.add(v), e.copy(this.target).add(d), this.object.lookAt(this.target), this.dispatchEvent(x), _ = 1, v.set(0, 0, 0);
	}, this.reset = function() {
		b = y.NONE, this.target.copy(this.target0), this.object.position.copy(this.position0), this.update();
	};
	function w() {
		return 2 * Math.PI / 60 / 60 * r.autoRotateSpeed;
	}
	function T() {
		return .95 ** r.zoomSpeed;
	}
	function E(e) {
		if (r.enabled !== !1) {
			if (e.preventDefault(), e.button === 0) {
				if (r.noRotate === !0) return;
				b = y.ROTATE, i.set(e.clientX, e.clientY);
			} else if (e.button === 1) {
				if (r.noZoom === !0) return;
				b = y.DOLLY, f.set(e.clientX, e.clientY);
			} else if (e.button === 2) {
				if (r.noPan === !0) return;
				b = y.PAN, s.set(e.clientX, e.clientY);
			}
			r.domElement.addEventListener("mousemove", D, !1), r.domElement.addEventListener("mouseup", O, !1), r.dispatchEvent(S);
		}
	}
	function D(e) {
		if (r.enabled !== !1) {
			e.preventDefault();
			var t = r.domElement === document ? r.domElement.body : r.domElement;
			if (b === y.ROTATE) {
				if (r.noRotate === !0) return;
				a.set(e.clientX, e.clientY), o.subVectors(a, i), r.rotateLeft(2 * Math.PI * o.x / t.clientWidth * r.rotateSpeed), r.rotateUp(2 * Math.PI * o.y / t.clientHeight * r.rotateSpeed), i.copy(a);
			} else if (b === y.DOLLY) {
				if (r.noZoom === !0) return;
				p.set(e.clientX, e.clientY), m.subVectors(p, f), m.y > 0 ? r.dollyIn() : r.dollyOut(), f.copy(p);
			} else if (b === y.PAN) {
				if (r.noPan === !0) return;
				c.set(e.clientX, e.clientY), l.subVectors(c, s), r.pan(l.x, l.y), s.copy(c);
			}
			r.update();
		}
	}
	function O() {
		r.enabled !== !1 && (r.domElement.removeEventListener("mousemove", D, !1), r.domElement.removeEventListener("mouseup", O, !1), r.dispatchEvent(C), b = y.NONE);
	}
	function k(e) {
		if (!(r.enabled === !1 || r.noZoom === !0)) {
			e.preventDefault();
			var t = 0;
			e.wheelDelta === void 0 ? e.detail !== void 0 && (t = -e.detail) : t = e.wheelDelta, t > 0 ? r.dollyOut() : r.dollyIn(), r.update(), r.dispatchEvent(S), r.dispatchEvent(C);
		}
	}
	function A(e) {
		if (!(r.enabled === !1 || r.noKeys === !0 || r.noPan === !0)) switch (e.keyCode) {
			case r.keys.UP:
				r.pan(0, r.keyPanSpeed), r.update();
				break;
			case r.keys.BOTTOM:
				r.pan(0, -r.keyPanSpeed), r.update();
				break;
			case r.keys.LEFT:
				r.pan(r.keyPanSpeed, 0), r.update();
				break;
			case r.keys.RIGHT:
				r.pan(-r.keyPanSpeed, 0), r.update();
				break;
		}
	}
	function j(e) {
		if (r.enabled !== !1) {
			switch (e.touches.length) {
				case 1:
					if (r.noRotate === !0) return;
					b = y.TOUCH_ROTATE, i.set(e.touches[0].pageX, e.touches[0].pageY);
					break;
				case 2:
					if (r.noZoom === !0) return;
					b = y.TOUCH_DOLLY;
					var t = e.touches[0].pageX - e.touches[1].pageX, n = e.touches[0].pageY - e.touches[1].pageY, a = Math.sqrt(t * t + n * n);
					f.set(0, a);
					break;
				case 3:
					if (r.noPan === !0) return;
					b = y.TOUCH_PAN, s.set(e.touches[0].pageX, e.touches[0].pageY);
					break;
				default: b = y.NONE;
			}
			r.dispatchEvent(S);
		}
	}
	function M(e) {
		if (r.enabled !== !1) {
			e.preventDefault(), e.stopPropagation();
			var t = r.domElement === document ? r.domElement.body : r.domElement;
			switch (e.touches.length) {
				case 1:
					if (r.noRotate === !0 || b !== y.TOUCH_ROTATE) return;
					a.set(e.touches[0].pageX, e.touches[0].pageY), o.subVectors(a, i), r.rotateLeft(2 * Math.PI * o.x / t.clientWidth * r.rotateSpeed), r.rotateUp(2 * Math.PI * o.y / t.clientHeight * r.rotateSpeed), i.copy(a), r.update();
					break;
				case 2:
					if (r.noZoom === !0 || b !== y.TOUCH_DOLLY) return;
					var n = e.touches[0].pageX - e.touches[1].pageX, u = e.touches[0].pageY - e.touches[1].pageY, d = Math.sqrt(n * n + u * u);
					p.set(0, d), m.subVectors(p, f), m.y > 0 ? r.dollyOut() : r.dollyIn(), f.copy(p), r.update();
					break;
				case 3:
					if (r.noPan === !0 || b !== y.TOUCH_PAN) return;
					c.set(e.touches[0].pageX, e.touches[0].pageY), l.subVectors(c, s), r.pan(l.x, l.y), s.copy(c), r.update();
					break;
				default: b = y.NONE;
			}
		}
	}
	function N() {
		r.enabled !== !1 && (r.dispatchEvent(C), b = y.NONE);
	}
	this.domElement.addEventListener("contextmenu", function(e) {
		e.preventDefault();
	}, !1), this.domElement.addEventListener("mousedown", E, !1), this.domElement.addEventListener("mousewheel", k, !1), this.domElement.addEventListener("DOMMouseScroll", k, !1), this.domElement.addEventListener("touchstart", j, !1), this.domElement.addEventListener("touchend", N, !1), this.domElement.addEventListener("touchmove", M, !1), window.addEventListener("keydown", A, !1);
}
OrbitControls.prototype = Object.create(THREE.EventDispatcher.prototype);
var round10_default = (e, t) => t === void 0 || +t == 0 ? Math.round(e) : (e = +e, t = +t, isNaN(e) || !(typeof t == "number" && t % 1 == 0) ? NaN : (e = e.toString().split("e"), e = Math.round(+(e[0] + "e" + (e[1] ? +e[1] - t : -t))), e = e.toString().split("e"), +(e[0] + "e" + (e[1] ? +e[1] + t : t)))), bspline_default = (e, t, n, r, i) => {
	let a = n.length, o = n[0].length;
	if (e < 0 || e > 1) throw Error("t out of bounds [0,1]: " + e);
	if (t < 1) throw Error("degree must be at least 1 (linear)");
	if (t > a - 1) throw Error("degree must be less than or equal to point count - 1");
	if (!i) {
		i = [];
		for (let e = 0; e < a; e++) i[e] = 1;
	}
	if (r) {
		if (r.length !== a + t + 1) throw Error("bad knot vector length");
	} else {
		r = [];
		for (let e = 0; e < a + t + 1; e++) r[e] = e;
	}
	let s = [t, r.length - 1 - t], l = r[s[0]], u = r[s[1]];
	e = e * (u - l) + l, e = Math.max(e, l), e = Math.min(e, u);
	let d;
	for (d = s[0]; d < s[1] && !(e >= r[d] && e <= r[d + 1]); d++);
	let f = [];
	for (let e = 0; e < a; e++) {
		f[e] = [];
		for (let t = 0; t < o; t++) f[e][t] = n[e][t] * i[e];
		f[e][o] = i[e];
	}
	let p;
	for (let n = 1; n <= t + 1; n++) for (let i = d; i > d - t - 1 + n; i--) {
		p = (e - r[i]) / (r[i + t + 1 - n] - r[i]);
		for (let e = 0; e < o + 1; e++) f[i][e] = (1 - p) * f[i - 1][e] + p * f[i][e];
	}
	let m = [];
	for (let e = 0; e < o; e++) m[e] = round10_default(f[d][e] / f[d][o], -9);
	return m;
}, THREEx = { Math: {} };
THREEx.Math.angle2 = function(t, n) {
	var r = new THREE.Vector2(t.x, t.y), i = new THREE.Vector2(n.x, n.y);
	return i.sub(r), i.normalize(), i.y < 0 ? -Math.acos(i.x) : Math.acos(i.x);
}, THREEx.Math.polar = function(e, t, n) {
	var r = {};
	return r.x = e.x + t * Math.cos(n), r.y = e.y + t * Math.sin(n), r;
};
function getBulgeCurvePoints(t, n, r, i) {
	var a, o, s, c, l, d, f, p, m, h = {};
	h.startPoint = c = t ? new THREE.Vector2(t.x, t.y) : new THREE.Vector2(0, 0), h.endPoint = l = n ? new THREE.Vector2(n.x, n.y) : new THREE.Vector2(1, 0), h.bulge = r ||= 1, d = 4 * Math.atan(r), f = c.distanceTo(l) / 2 / Math.sin(d / 2), s = THREEx.Math.polar(t, f, THREEx.Math.angle2(c, l) + (Math.PI / 2 - d / 2)), h.segments = i ||= Math.max(Math.abs(Math.ceil(d / (Math.PI / 18))), 6), p = THREEx.Math.angle2(s, c), m = d / i;
	var g = [];
	for (g.push(new THREE.Vector3(c.x, c.y, 0)), o = 1; o <= i - 1; o++) a = THREEx.Math.polar(s, Math.abs(f), p + m * o), g.push(new THREE.Vector3(a.x, a.y, 0));
	return g;
}
function Viewer(c, u, f, p, m) {
	q(c);
	var h = new THREE.Scene(), g, _, v, y = {
		min: {
			x: 0,
			y: 0,
			z: 0
		},
		max: {
			x: 0,
			y: 0,
			z: 0
		}
	};
	for (g = 0; g < c.entities.length; g++) {
		if (_ = c.entities[g], v = j(_, c), v) {
			var b = new THREE.Box3().setFromObject(v);
			isFinite(b.min.x) && y.min.x > b.min.x && (y.min.x = b.min.x), isFinite(b.min.y) && y.min.y > b.min.y && (y.min.y = b.min.y), isFinite(b.min.z) && y.min.z > b.min.z && (y.min.z = b.min.z), isFinite(b.max.x) && y.max.x < b.max.x && (y.max.x = b.max.x), isFinite(b.max.y) && y.max.y < b.max.y && (y.max.y = b.max.y), isFinite(b.max.z) && y.max.z < b.max.z && (y.max.z = b.max.z), h.add(v);
		}
		v = null;
	}
	f ||= u.clientWidth, p ||= u.clientHeight;
	var x = f / p, S = {
		x: y.max.x,
		y: y.max.y
	}, C = {
		x: y.min.x,
		y: y.min.y
	}, w = S.x - C.x, T = S.y - C.y, E = E || {
		x: w / 2 + C.x,
		y: T / 2 + C.y
	};
	x > Math.abs(w / T) ? w = T * x : T = w / x;
	var D = {
		bottom: -T / 2,
		left: -w / 2,
		top: T / 2,
		right: w / 2,
		center: {
			x: E.x,
			y: E.y
		}
	}, O = new THREE.OrthographicCamera(D.left, D.right, D.top, D.bottom, 1, 19);
	O.position.z = 10, O.position.x = D.center.x, O.position.y = D.center.y;
	var k = this.renderer = new THREE.WebGLRenderer();
	k.setSize(f, p), k.setClearColor(268435455, 1), u.appendChild(k.domElement), u.style.display = "block";
	var A = new OrbitControls(O, u);
	A.target.x = O.position.x, A.target.y = O.position.y, A.target.z = 0, A.zoomSpeed = 3, this.render = function() {
		k.render(h, O);
	}, A.addEventListener("change", this.render), this.render(), A.update(), this.resize = function(e, t) {
		var n = k.domElement.width, r = k.domElement.height, i = e / n, a = t / r;
		O.top = a * O.top, O.bottom = a * O.bottom, O.left = i * O.left, O.right = i * O.right, k.setSize(e, t), k.setClearColor(268435455, 1), this.render();
	};
	function j(e, t) {
		var n;
		if (e.type === "CIRCLE" || e.type === "ARC") n = z(e, t);
		else if (e.type === "LWPOLYLINE" || e.type === "LINE" || e.type === "POLYLINE") n = R(e, t);
		else if (e.type === "TEXT") n = H(e, t);
		else if (e.type === "SOLID") n = V(e, t);
		else if (e.type === "POINT") n = U(e, t);
		else if (e.type === "INSERT") n = G(e, t);
		else if (e.type === "SPLINE") n = I(e, t);
		else if (e.type === "MTEXT") n = N(e, t);
		else if (e.type === "ELLIPSE") n = M(e, t);
		else if (e.type === "DIMENSION") {
			var r = e.dimensionType & 7;
			r === 0 ? n = W(e, t) : console.log("Unsupported Dimension type: " + r);
		} else console.log("Unsupported Entity Type: " + e.type);
		return n;
	}
	function M(t, n) {
		var r = K(t, n), i = Math.sqrt(t.majorAxisEndPoint.x ** 2 + t.majorAxisEndPoint.y ** 2), a = i * t.axisRatio, o = Math.atan2(t.majorAxisEndPoint.y, t.majorAxisEndPoint.x), s = new THREE.EllipseCurve(t.center.x, t.center.y, i, a, t.startAngle, t.endAngle, !1, o).getPoints(50), c = new THREE.BufferGeometry().setFromPoints(s), l = new THREE.LineBasicMaterial({
			linewidth: 1,
			color: r
		});
		return new THREE.Line(c, l);
	}
	function N(t, n) {
		var r = K(t, n);
		if (!m) return console.log("font parameter not set. Ignoring text entity.");
		var i = P(parseDxfMTextContent(t.text), t, r), a = F(i.text, i.style, t, r);
		if (!a) return null;
		var s = new THREE.Object3D();
		return s.add(a), s;
	}
	function P(e, t, n) {
		let r = {
			horizontalAlignment: "left",
			textHeight: t.height
		};
		var i = [];
		for (let o of e) if (typeof o == "string") o.startsWith("pxq") && o.endsWith(";") ? o.indexOf("c") === -1 ? o.indexOf("l") === -1 ? o.indexOf("r") === -1 ? o.indexOf("j") !== -1 && (r.horizontalAlignment = "justify") : r.horizontalAlignment = "right" : r.horizontalAlignment = "left" : r.horizontalAlignment = "center" : i.push(o);
		else if (Array.isArray(o)) {
			var a = P(o, t, n);
			i.push(a.text);
		} else typeof o == "object" && o.S && o.S.length === 3 && i.push(o.S[0] + "/" + o.S[2]);
		return {
			text: i.join(),
			style: r
		};
	}
	function F(t, n, r, i) {
		if (!t) return null;
		let o = new Text();
		if (o.text = decodeUnicodeEscapes(t.replaceAll("\\P", "\n").replaceAll("\\X", "\n")), o.font = m, o.fontSize = n.textHeight, o.maxWidth = r.width, o.position.x = r.position.x, o.position.y = r.position.y, o.position.z = r.position.z, o.textAlign = n.horizontalAlignment, o.color = i, r.rotation && (o.rotation.z = r.rotation * Math.PI / 180), r.directionVector) {
			var s = r.directionVector;
			o.rotation.z = new THREE.Vector3(1, 0, 0).angleTo(new THREE.Vector3(s.x, s.y, s.z));
		}
		switch (r.attachmentPoint) {
			case 1:
				o.anchorX = "left", o.anchorY = "top";
				break;
			case 2:
				o.anchorX = "center", o.anchorY = "top";
				break;
			case 3:
				o.anchorX = "right", o.anchorY = "top";
				break;
			case 4:
				o.anchorX = "left", o.anchorY = "middle";
				break;
			case 5:
				o.anchorX = "center", o.anchorY = "middle";
				break;
			case 6:
				o.anchorX = "right", o.anchorY = "middle";
				break;
			case 7:
				o.anchorX = "left", o.anchorY = "bottom";
				break;
			case 8:
				o.anchorX = "center", o.anchorY = "bottom";
				break;
			case 9:
				o.anchorX = "right", o.anchorY = "bottom";
				break;
			default: return;
		}
		return o.sync(() => {
			if (o.textAlign !== "left") {
				o.geometry.computeBoundingBox();
				var e = o.geometry.boundingBox.max.x - o.geometry.boundingBox.min.x;
				o.textAlign === "center" && (o.position.x += (r.width - e) / 2), o.textAlign === "right" && (o.position.x += r.width - e);
			}
			k.render(h, O);
		}), o;
	}
	function I(t, n) {
		var r = K(t, n), i = L(t.controlPoints, t.degreeOfSplineCurve, t.knotValues, 100), a = new THREE.BufferGeometry().setFromPoints(i), o = new THREE.LineBasicMaterial({
			linewidth: 1,
			color: r
		});
		return new THREE.Line(a, o);
	}
	function L(t, n, r, i, a) {
		let o = [], s = t.map(function(e) {
			return [e.x, e.y];
		}), c = [r[n]], u = [r[n], r[r.length - 1 - n]];
		for (let e = n + 1; e < r.length - n; ++e) c[c.length - 1] !== r[e] && c.push(r[e]);
		i ||= 25;
		for (let t = 1; t < c.length; ++t) {
			let d = c[t - 1], f = c[t];
			for (let t = 0; t <= i; ++t) {
				let c = (t / i * (f - d) + d - u[0]) / (u[1] - u[0]);
				c = Math.max(c, 0), c = Math.min(c, 1);
				let p = bspline_default(c, n, s, r, a);
				o.push(new THREE.Vector2(p[0], p[1]));
			}
		}
		return o;
	}
	function R(n, r) {
		let i = [], a = K(n, r);
		var o, s, c, l, u, f, p, m;
		if (!n.vertices) return console.log("entity missing vertices.");
		for (p = 0; p < n.vertices.length; p++) if (n.vertices[p].bulge) {
			f = n.vertices[p].bulge, l = n.vertices[p], u = p + 1 < n.vertices.length ? n.vertices[p + 1] : i[0];
			let e = getBulgeCurvePoints(l, u, f);
			i.push.apply(i, e);
		} else c = n.vertices[p], i.push(new THREE.Vector3(c.x, c.y, 0));
		n.shape && i.push(i[0]), n.lineType && (s = r.tables.lineType.lineTypes[n.lineType]), o = s && s.pattern && s.pattern.length !== 0 ? new THREE.LineDashedMaterial({
			color: a,
			gapSize: 4,
			dashSize: 4
		}) : new THREE.LineBasicMaterial({
			linewidth: 1,
			color: a
		});
		var h = new BufferGeometry().setFromPoints(i);
		return m = new THREE.Line(h, o), m;
	}
	function z(t, n) {
		var r, i;
		t.type === "CIRCLE" ? (r = t.startAngle || 0, i = r + 2 * Math.PI) : (r = t.startAngle, i = t.endAngle);
		var a = new THREE.ArcCurve(0, 0, t.radius, r, i).getPoints(32), o = new THREE.BufferGeometry().setFromPoints(a), s = new THREE.LineBasicMaterial({ color: K(t, n) }), c = new THREE.Line(o, s);
		return c.position.x = t.center.x, c.position.y = t.center.y, c.position.z = t.center.z, c;
	}
	function B(e, t, n, r) {
		var a = new Vector3(), o = new Vector3();
		a.subVectors(n, t), o.subVectors(r, t), a.cross(o);
		var s = new Vector3(t.x, t.y, t.z), c = new Vector3(n.x, n.y, n.z), l = new Vector3(r.x, r.y, r.z);
		a.z < 0 ? e.push(l, c, s) : e.push(s, c, l);
	}
	function V(t, n) {
		var r, i, a = new THREE.BufferGeometry(), o = t.points;
		return i = [], B(i, o[0], o[1], o[2]), B(i, o[1], o[2], o[3]), r = new THREE.MeshBasicMaterial({ color: K(t, n) }), a.setFromPoints(i), new THREE.Mesh(a, r);
	}
	function H(t, n) {
		if (!m) return console.log("font parameter not set. Ignoring text entity.");
		if (!t.text) return null;
		var r = new Text();
		r.text = decodeUnicodeEscapes(
			t.text.replaceAll("\\P", "\n").replaceAll("\\X", "\n")
		);
		r.font = m;
		r.fontSize = t.textHeight || 12;
		r.whiteSpace = "nowrap";
		r.position.x = t.startPoint.x;
		r.position.y = t.startPoint.y;
		r.position.z = t.startPoint.z;
		r.color = K(t, n);
		if (t.rotation) r.rotation.z = t.rotation * Math.PI / 180;

		var i = t.halign ?? 0;
		var a = t.valign ?? 0;

		switch (i) {
			case 1:
				r.anchorX = "center";
				r.textAlign = "center";
				break;
			case 2:
				r.anchorX = "right";
				r.textAlign = "right";
				break;
			case 4:
				r.anchorX = "center";
				r.textAlign = "center";
				break;
			case 3:
			case 5:
				r.anchorX = "left";
				r.textAlign = "left";
				break;
			default:
				r.anchorX = "left";
				r.textAlign = "left";
		}

		switch (a) {
			case 1:
				r.anchorY = "bottom";
				break;
			case 2:
				r.anchorY = "middle";
				break;
			case 3:
				r.anchorY = "top";
				break;
			default:
				r.anchorY = "bottom-baseline";
		}

		if (t.endPoint && (i === 3 || i === 5)) {
			var o = t.endPoint.x - t.startPoint.x;
			var s = t.endPoint.y - t.startPoint.y;
			var c = Math.hypot(o, s);
			if (c > 0) r.maxWidth = c;
			if (!t.rotation) r.rotation.z = Math.atan2(s, o);
		}

		r.sync(() => {
			k.render(h, O);
		});
		return r;
	}
	function U(t, i) {
		var a = new THREE.BufferGeometry(), o, s;
		a.setAttribute("position", new Float32BufferAttribute([
			t.position.x,
			t.position.y,
			t.position.z
		], 3));
		var c = K(t, i);
		o = new THREE.PointsMaterial({
			size: .1,
			color: new Color(c)
		}), s = new THREE.Points(a, o), h.add(s);
	}
	function W(t, n) {
		var r = n.blocks[t.block];
		if (!r || !r.entities) return null;
		for (var i = new THREE.Object3D(), a = 0; a < r.entities.length; a++) {
			var o = j(r.entities[a], n, i);
			o && i.add(o);
		}
		return i;
	}
	function G(t, n) {
		var r = n.blocks[t.name];
		if (!r.entities) return null;
		var i = new THREE.Object3D();
		t.xScale && (i.scale.x = t.xScale), t.yScale && (i.scale.y = t.yScale), t.rotation && (i.rotation.z = t.rotation * Math.PI / 180), t.position && (i.position.x = t.position.x, i.position.y = t.position.y, i.position.z = t.position.z);
		for (var a = 0; a < r.entities.length; a++) {
			var o = j(r.entities[a], n, i);
			o && i.add(o);
		}
		return i;
	}
	function K(e, t) {
		var n = 0;
		return e.color ? n = e.color : t.tables && t.tables.layer && t.tables.layer.layers[e.layer] && (n = t.tables.layer.layers[e.layer].color), (n == null || n === 16777215) && (n = 0), n;
	}
	function q(e) {
		var t, n;
		if (!(!e.tables || !e.tables.lineType)) {
			var r = e.tables.lineType.lineTypes;
			for (n in r) t = r[n], t.pattern && (t.material = J(t.pattern));
		}
	}
	function J(t) {
		var n, r = {}, i = 0;
		for (n = 0; n < t.length; n++) i += Math.abs(t[n]);
		return r.uniforms = THREE.UniformsUtils.merge([
			THREE.UniformsLib.common,
			THREE.UniformsLib.fog,
			{
				pattern: {
					type: "fv1",
					value: t
				},
				patternLength: {
					type: "f",
					value: i
				}
			}
		]), r.vertexShader = [
			"attribute float lineDistance;",
			"varying float vLineDistance;",
			THREE.ShaderChunk.color_pars_vertex,
			"void main() {",
			THREE.ShaderChunk.color_vertex,
			"vLineDistance = lineDistance;",
			"gl_Position = projectionMatrix * modelViewMatrix * vec4( position, 1.0 );",
			"}"
		].join("\n"), r.fragmentShader = [
			"uniform vec3 diffuse;",
			"uniform float opacity;",
			"uniform float pattern[" + t.length + "];",
			"uniform float patternLength;",
			"varying float vLineDistance;",
			THREE.ShaderChunk.color_pars_fragment,
			THREE.ShaderChunk.fog_pars_fragment,
			"void main() {",
			"float pos = mod(vLineDistance, patternLength);",
			"for ( int i = 0; i < " + t.length + "; i++ ) {",
			"pos = pos - abs(pattern[i]);",
			"if( pos < 0.0 ) {",
			"if( pattern[i] > 0.0 ) {",
			"gl_FragColor = vec4(1.0, 0.0, 0.0, opacity );",
			"break;",
			"}",
			"discard;",
			"}",
			"}",
			THREE.ShaderChunk.color_fragment,
			THREE.ShaderChunk.fog_fragment,
			"}"
		].join("\n"), r;
	}
}
export { Viewer };
