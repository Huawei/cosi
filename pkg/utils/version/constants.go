/*
Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
  http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package version used to set and clean the service version
package version

const (
	versionConfigMapName = "huawei-cosi-version"
)

var (
	buildVersion string
	buildArch    string
)

var (
	// OSArch the architecture of service
	OSArch = buildArch

	// COSIDriverVersion the version of huawei cosi driver
	COSIDriverVersion = buildVersion

	// LivenessProbeVersion the version of livenessprobe
	LivenessProbeVersion = buildVersion
)
