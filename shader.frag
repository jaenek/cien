#version 410 core
uniform vec2 v2Resolution;
uniform float fGlobalTime;

out vec4 out_color;

// takes a position in space and returns a distance from surface
float map(vec3 pos) {
	float d1 = length(max(abs(pos)-0.2, 0.0))- 0.05;
	return d1;
}

// takes a position and returns an aproximate normal of this position
vec3 normal(vec3 pos) {
	vec2 e = vec2(0.0001, 0.0);
	return normalize(vec3(map(pos+e.xyy)-map(pos-e.xyy),
	                      map(pos+e.yxy)-map(pos-e.yxy),
	                      map(pos+e.yyx)-map(pos-e.yyx)));
}

void main() {
	vec2 p = (2.0*gl_FragCoord.xy-v2Resolution.xy)/v2Resolution.y;

	float an = 0.4*fGlobalTime; // changing angle

	vec3 ro = vec3(1.0*sin(an), 1.0*sin(an), 1.0*cos(an)); // camera position
	vec3 ta = vec3(0.0, 0.0, 0.0); // target position

	vec3 ww = normalize(ta - ro);
	vec3 uu = normalize(cross(ww, vec3(0,1,0)));
	vec3 vv = normalize(cross(uu,ww));

	vec3 rd = normalize(p.x*uu + p.y*vv + 1.5*ww); // pixel we are looking at

	vec3 col = vec3(p.x);

	float t = 0.0;
	for (int i = 0; i < 100; i++) {
		vec3 pos = ro + t*rd; // step from ro to rd in t interval

		float h = map(pos);
		if (h < 0.001) {
			break; // break the loop if close to surface or behind it
		} else if (t > 20.0) {
			break; // far clipping plane
		}

		t += h;
	}

	// if the ray hit something color the pixel appropriately
	if (t < 20.0) {
		vec3 pos = ro + t*rd;
		vec3 nor = normal(pos);

		vec3 key_dir = normalize(vec3(0.8, 0.4, 0.2));
		float key_dif = clamp(dot(nor, key_dir), 0.0, 1.0);
		col = vec3(1.0)*key_dif;
	}

	out_color = vec4(col, 1);
}

