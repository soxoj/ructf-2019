#version 450
#extension GL_ARB_separate_shader_objects : enable

layout (location = 0) in vec3 vertexPos;
layout (location = 1) in uint instanceId;

uniform mat4 viewProjMatrix;
uniform ivec4 numBuildings;
uniform vec4 buildingSize;

void main()
{
	int xiOffset = int(instanceId) % numBuildings.x;
	int ziOffset = int(instanceId) / numBuildings.x;

	float xOffset = float(xiOffset) * (buildingSize.x + buildingSize.y) + buildingSize.x * 0.5f + buildingSize.y;
	float zOffset = float(ziOffset) * (buildingSize.x + buildingSize.y) + buildingSize.x * 0.5f + buildingSize.y;

	vec4 worldPos = vec4(vertexPos * buildingSize.x, 1.0f);
	worldPos.xz += vec2(xOffset, zOffset);

	gl_Position = worldPos;
}
