// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Xiang Li
// Create: 2022-05-16
// Description: offline pod cache limit directory init function

package cachelimit

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"isula.org/rubik/pkg/checkpoint"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
	"isula.org/rubik/pkg/perf"
	"isula.org/rubik/pkg/try"
	"isula.org/rubik/pkg/typedef"
)

// TestGetNUMANum testcase
func TestGetNUMANum(t *testing.T) {
	threeNodeDir := try.GenTestDir().String()
	for i := 0; i < 3; i++ {
		nodeDir := filepath.Join(threeNodeDir, fmt.Sprintf("node%d", i))
		try.MkdirAll(nodeDir, constant.DefaultDirMode)
	}

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
		compare bool
	}{
		{
			name:    "TC-right numa folder",
			args:    args{path: numaNodeDir},
			wantErr: false,
			compare: false,
		},
		{
			name:    "TC-three numa foler",
			args:    args{path: threeNodeDir},
			want:    3,
			wantErr: false,
			compare: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getNUMANum(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNUMANum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.compare {
				if got != tt.want {
					t.Errorf("getNUMANum() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// TestGetBinaryMask testcase
func TestGetBinaryMask(t *testing.T) {
	file7ff := filepath.Join(try.GenTestDir().String(), "7ff")
	file3ff := filepath.Join(try.GenTestDir().String(), "3ff")
	fileNotHex := filepath.Join(try.GenTestDir().String(), "nohex")

	tests := []struct {
		preHook func(t *testing.T)
		name    string
		path    string
		want    int
		wantErr bool
	}{
		{
			name:    "TC-7ff",
			path:    file7ff,
			want:    11,
			wantErr: false,
			preHook: func(t *testing.T) {
				try.WriteFile(file7ff, []byte("7ff"), constant.DefaultFileMode)
			},
		},
		{
			name:    "TC-3ff",
			path:    file3ff,
			want:    10,
			wantErr: false,
			preHook: func(t *testing.T) {
				try.WriteFile(file3ff, []byte("3ff"), constant.DefaultFileMode)
			},
		},
		{
			name:    "TC-not hex format",
			path:    fileNotHex,
			wantErr: true,
			preHook: func(t *testing.T) {
				try.WriteFile(fileNotHex, []byte("ghi"), constant.DefaultFileMode)
			},
		},
		{
			name:    "TC-file not exist",
			path:    "/file/not/exist",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preHook != nil {
				tt.preHook(t)
			}
			got, err := getBinaryMask(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("getBinaryMask() error = %v, wantErr %v, file = %v", err, tt.wantErr, tt.path)
				return
			}
			if err == nil {
				if got != tt.want {
					t.Errorf("getBinaryMask() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// TestCalcLimitedCacheValue testcase
func TestCalcLimitedCacheValue(t *testing.T) {
	testFile := filepath.Join(try.GenTestDir().String(), "testFile")
	type fields struct {
		level     string
		L3Percent int
		MbPercent int
	}
	type args struct {
		path string
	}
	tests := []struct {
		preHook  func(t *testing.T)
		postHook func(t *testing.T)
		name     string
		fields   fields
		args     args
		want     string
		wantErr  bool
	}{
		{
			name: "TC-7ff",
			args: args{testFile},
			want: "1",
			fields: fields{
				L3Percent: 10,
				MbPercent: 10,
			},
			preHook: func(t *testing.T) {
				try.WriteFile(testFile, []byte("7ff"), constant.DefaultFileMode)
			},
		},
		{
			name: "TC-fffff",
			args: args{testFile},
			want: "3",
			fields: fields{
				L3Percent: 10,
				MbPercent: 10,
			},
			preHook: func(t *testing.T) {
				try.WriteFile(testFile, []byte("fffff"), constant.DefaultFileMode)
			},
		},
		{
			name: "TC-ff",
			args: args{testFile},
			want: "1",
			fields: fields{
				L3Percent: 10,
			},
			preHook: func(t *testing.T) {
				try.WriteFile(testFile, []byte("ff"), constant.DefaultFileMode)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clSet := &cacheLimitSet{
				level:     tt.fields.level,
				L3Percent: tt.fields.L3Percent,
				MbPercent: tt.fields.MbPercent,
			}
			if tt.preHook != nil {
				tt.preHook(t)
			}
			got, err := calcLimitedCacheValue(tt.args.path, clSet.L3Percent)
			if (err != nil) != tt.wantErr {
				t.Errorf("cacheLimitSet.calcLimitedCacheValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("cacheLimitSet.calcLimitedCacheValue() = %v, want %v", got, tt.want)
			}
			if tt.postHook != nil {
				tt.postHook(t)
			}
		})
	}
}

// TestWriteResctrlSchemata testcase
func TestWriteResctrlSchemata(t *testing.T) {
	testFolder := try.GenTestDir().String()
	assert.NoError(t, setMaskFile(t, testFolder, "3ff"))
	type fields struct {
		level     string
		clDir     string
		L3Percent int
		MbPercent int
	}
	type args struct {
		llc     string
		numaNum int
	}
	tests := []struct {
		preHook  func(t *testing.T)
		postHook func(t *testing.T)
		name     string
		fields   fields
		args     args
		wantErr  bool
	}{
		{
			name: "TC-normal",
			fields: fields{
				level:     lowLevel,
				clDir:     filepath.Join(testFolder, "normal"),
				L3Percent: 30,
				MbPercent: 30,
			},
			args:    args{llc: "3ff", numaNum: 2},
			wantErr: false,
		},
		{
			name: "TC-cache limit dir not set",
			fields: fields{
				level:     lowLevel,
				L3Percent: 30,
				MbPercent: 30,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clSet := &cacheLimitSet{
				level:     tt.fields.level,
				clDir:     tt.fields.clDir,
				L3Percent: tt.fields.L3Percent,
				MbPercent: tt.fields.MbPercent,
			}
			if tt.preHook != nil {
				tt.preHook(t)
			}
			if err := clSet.writeResctrlSchemata(tt.args.numaNum); (err != nil) != tt.wantErr {
				t.Errorf("cacheLimitSet.writeResctrlSchemata() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.postHook != nil {
				tt.postHook(t)
			}
		})
	}
}

// TestCheckCacheCfg testcase
func TestCheckCacheCfg(t *testing.T) {
	type args struct {
		cfg config.CacheConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wantMsg string
	}{
		{
			name: "TC-static mode config",
			args: args{cfg: config.CacheConfig{
				DefaultLimitMode: staticMode,
				AdjustInterval:   minAdjustInterval + 1,
				PerfDuration:     minPerfDur + 1,
				L3Percent: config.MultiLvlPercent{
					Low:  minPercent + 1,
					Mid:  maxPercent/2 + 1,
					High: maxPercent - 1,
				},
				MemBandPercent: config.MultiLvlPercent{
					Low:  minPercent + 1,
					Mid:  maxPercent/2 + 1,
					High: maxPercent - 1,
				},
			}},
		},
		{
			name: "TC-invalid mode config",
			args: args{cfg: config.CacheConfig{
				DefaultLimitMode: "invalid mode",
			}},
			wantErr: true,
			wantMsg: dynamicMode,
		},
		{
			name: "TC-invalid adjust interval less than min value",
			args: args{cfg: config.CacheConfig{
				DefaultLimitMode: staticMode,
				AdjustInterval:   minAdjustInterval - 1,
			}},
			wantErr: true,
			wantMsg: strconv.Itoa(minAdjustInterval),
		},
		{
			name: "TC-invalid adjust interval greater than max value",
			args: args{cfg: config.CacheConfig{
				DefaultLimitMode: staticMode,
				AdjustInterval:   maxAdjustInterval + 1,
			}},
			wantErr: true,
			wantMsg: strconv.Itoa(maxAdjustInterval),
		},
		{
			name: "TC-invalid perf duration less than min value",
			args: args{cfg: config.CacheConfig{
				DefaultLimitMode: staticMode,
				AdjustInterval:   maxAdjustInterval/2 + 1,
				PerfDuration:     minPerfDur - 1,
			}},
			wantErr: true,
			wantMsg: strconv.Itoa(minPerfDur),
		},
		{
			name: "TC-invalid perf duration greater than max value",
			args: args{cfg: config.CacheConfig{
				DefaultLimitMode: staticMode,
				AdjustInterval:   maxAdjustInterval/2 + 1,
				PerfDuration:     maxPerfDur + 1,
			}},
			wantErr: true,
			wantMsg: strconv.Itoa(maxPerfDur),
		},
		{
			name: "TC-invalid percent value",
			args: args{cfg: config.CacheConfig{
				DefaultLimitMode: staticMode,
				AdjustInterval:   maxAdjustInterval/2 + 1,
				PerfDuration:     maxPerfDur/2 + 1,
				L3Percent: config.MultiLvlPercent{
					Low: minPercent - 1,
				},
			}},
			wantErr: true,
			wantMsg: strconv.Itoa(minPercent),
		},
		{
			name: "TC-invalid l3 percent low value larger than mid value",
			args: args{cfg: config.CacheConfig{
				DefaultLimitMode: staticMode,
				AdjustInterval:   maxAdjustInterval/2 + 1,
				PerfDuration:     maxPerfDur/2 + 1,
				L3Percent: config.MultiLvlPercent{
					Low:  minPercent + 2,
					Mid:  minPercent + 1,
					High: minPercent + 1,
				},
				MemBandPercent: config.MultiLvlPercent{
					Low:  minPercent,
					Mid:  minPercent + 1,
					High: minPercent + 2,
				},
			}},
			wantErr: true,
			wantMsg: "low<=mid<=high",
		},
		{
			name: "TC-invalid memband percent mid value larger than high value",
			args: args{cfg: config.CacheConfig{
				DefaultLimitMode: staticMode,
				AdjustInterval:   maxAdjustInterval/2 + 1,
				PerfDuration:     maxPerfDur/2 + 1,
				L3Percent: config.MultiLvlPercent{
					Low:  minPercent,
					Mid:  minPercent + 1,
					High: minPercent + 2,
				},
				MemBandPercent: config.MultiLvlPercent{
					Low:  minPercent,
					Mid:  maxPercent/2 + 1,
					High: maxPercent / 2,
				},
			}},
			wantErr: true,
			wantMsg: "low<=mid<=high",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkCacheCfg(&tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkCacheCfg() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("checkCacheCfg() error = %v, wantMsg %v", err, tt.wantMsg)
			}
		})
	}
}

func setMaskFile(t *testing.T, resctrlDir string, data string) error {
	maskDir := filepath.Join(resctrlDir, "info", "L3")
	maskFile := filepath.Join(maskDir, "cbm_mask")
	if err := os.MkdirAll(maskDir, constant.DefaultDirMode); err != nil {
		return err
	}
	if err := ioutil.WriteFile(maskFile, []byte(data), constant.DefaultFileMode); err != nil {
		return err
	}
	return nil
}

// TestInitCacheLimitDir testcase
func TestInitCacheLimitDir(t *testing.T) {
	resctrlDir := try.GenTestDir().String()
	type args struct {
		cfg config.CacheConfig
	}
	tests := []struct {
		setMaskFile func(t *testing.T) error
		name        string
		args        args
		wantErr     bool
	}{
		{
			name: "TC-valid cache limit dir setting",
			args: args{cfg: config.CacheConfig{
				DefaultResctrlDir: resctrlDir,
				DefaultLimitMode:  staticMode,
			}},
			setMaskFile: func(t *testing.T) error {
				return setMaskFile(t, resctrlDir, "3ff")
			},
		},
		{
			name: "TC-empty resctrl dir",
			args: args{config.CacheConfig{
				DefaultResctrlDir: "",
				DefaultLimitMode:  staticMode,
			}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setMaskFile != nil {
				assert.NoError(t, tt.setMaskFile(t))
			}
			if err := initCacheLimitDir(&tt.args.cfg); (err != nil) != tt.wantErr {
				t.Errorf("initCacheLimitDir() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSetClDir testcase
func TestSetClDir(t *testing.T) {
	testRoot := try.GenTestDir().String()
	_, err := os.Create(filepath.Join(testRoot, "test"))
	assert.NoError(t, err)
	type fields struct {
		level     string
		clDir     string
		L3Percent int
		MbPercent int
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:   "TC-normal cache limit dir",
			fields: fields{clDir: testRoot},
		},
		{
			name:    "TC-empty dir",
			wantErr: true,
		},
		{
			name:    "TC-path not exist",
			fields:  fields{clDir: "/path/not/exist"},
			wantErr: true,
		},
		{
			name:    "TC-path not exist",
			fields:  fields{clDir: filepath.Join(testRoot, "test", "test")},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clSet := &cacheLimitSet{
				level:     tt.fields.level,
				clDir:     tt.fields.clDir,
				L3Percent: tt.fields.L3Percent,
				MbPercent: tt.fields.MbPercent,
			}
			if err := clSet.setClDir(); (err != nil) != tt.wantErr {
				t.Errorf("cacheLimitSet.setClDir() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCheckResctrlExist testcase
func TestCheckResctrlExist(t *testing.T) {
	resctrlDir := try.GenTestDir().String()
	resctrlDirNoSchemataFile := try.GenTestDir().String()
	schemataPath := filepath.Join(resctrlDir, schemataFile)
	_, err := os.Create(schemataPath)
	assert.NoError(t, err)
	type args struct {
		cfg config.CacheConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TC-resctrl exist",
			args: args{cfg: config.CacheConfig{
				DefaultResctrlDir: resctrlDir,
			}},
		},
		{
			name: "TC-resctrl exist but not schemata file",
			args: args{cfg: config.CacheConfig{
				DefaultResctrlDir: resctrlDirNoSchemataFile,
			}},
			wantErr: true,
		},
		{
			name: "TC-resctrl not exist",
			args: args{cfg: config.CacheConfig{
				DefaultResctrlDir: "/path/not/exist",
			}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkResctrlExist(&tt.args.cfg); (err != nil) != tt.wantErr {
				t.Errorf("checkResctrlExist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDoFlush testcase
func TestAdjustCacheLimit(t *testing.T) {
	resctrlDir := try.GenTestDir().String()
	assert.NoError(t, setMaskFile(t, resctrlDir, "3ff"))

	type fields struct {
		level     string
		clDir     string
		L3Percent int
		MbPercent int
	}
	type args struct {
		clValue string
	}
	tests := []struct {
		preHook func(t *testing.T)
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TC-adjust success",
			fields: fields{
				level:     lowLevel,
				clDir:     filepath.Join(filepath.Clean(resctrlDir), dirPrefix+lowLevel),
				L3Percent: 10,
				MbPercent: 10,
			},
		},
		{
			name: "TC-l3PercentDynamic",
			fields: fields{
				level:     lowLevel,
				clDir:     filepath.Join(filepath.Clean(resctrlDir), dirPrefix+lowLevel),
				L3Percent: l3PercentDynamic,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clSet := &cacheLimitSet{
				level:     tt.fields.level,
				clDir:     tt.fields.clDir,
				L3Percent: tt.fields.L3Percent,
				MbPercent: tt.fields.MbPercent,
			}
			if tt.preHook != nil {
				tt.preHook(t)
			}
			if err := clSet.doFlush(); (err != nil) != tt.wantErr {
				t.Errorf("clSet.doFlush() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetPodCacheMiss(t *testing.T) {
	if !perf.HwSupport() {
		t.Skipf("%s only run on physical machine", t.Name())
	}
	testCGRoot := filepath.Join(config.CgroupRoot, "perf_event", t.Name())
	type fields struct {
		podID           string
		cgroupPath      string
		cacheLimitLevel string
		containers      map[string]*typedef.ContainerInfo
	}
	type args struct {
		cgroupRoot string
		perfDu     int
	}
	tests := []struct {
		preHook  func(t *testing.T)
		postHook func(t *testing.T)
		name     string
		fields   fields
		args     args
		want     int
	}{
		{
			name: "TC-get pod cache miss success",
			fields: fields{
				podID:           "abcd",
				cgroupPath:      t.Name(),
				cacheLimitLevel: lowLevel,
				containers:      make(map[string]*typedef.ContainerInfo),
			},
			preHook: func(t *testing.T) {
				try.MkdirAll(testCGRoot, constant.DefaultDirMode)
				try.WriteFile(filepath.Join(testCGRoot, "tasks"), []byte(fmt.Sprint(os.Getpid())), constant.DefaultFileMode)
			},
			postHook: func(t *testing.T) {
				try.WriteFile(filepath.Join(config.CgroupRoot, "perf_event", "tasks"), []byte(fmt.Sprint(os.Getpid())), constant.DefaultFileMode)
				try.RemoveAll(testCGRoot)
			},
			args: args{cgroupRoot: config.CgroupRoot, perfDu: 1},
		},
		{
			name: "TC-get pod cache miss failed",
			fields: fields{
				podID:           "abcd",
				cgroupPath:      t.Name(),
				cacheLimitLevel: middleLevel,
				containers:      make(map[string]*typedef.ContainerInfo),
			},
		},
	}
	for _, tt := range tests {
		name := t.Name()
		fmt.Println(name)
		t.Run(tt.name, func(t *testing.T) {
			p := &typedef.PodInfo{
				UID:             tt.fields.podID,
				CgroupPath:      tt.fields.cgroupPath,
				CacheLimitLevel: tt.fields.cacheLimitLevel,
				Containers:      tt.fields.containers,
			}
			if tt.preHook != nil {
				tt.preHook(t)
			}
			getPodCacheMiss(p, tt.args.perfDu)
			if tt.postHook != nil {
				tt.postHook(t)
			}
		})
	}
}

func TestStartDynamic(t *testing.T) {
	if !perf.HwSupport() {
		t.Skipf("%s only run on physical machine", t.Name())
	}
	initCpm()
	startDynamic(&config.CacheConfig{}, 0, 0)
	resctrlDir := try.GenTestDir().String()
	testCGRoot := filepath.Join(config.CgroupRoot, "perf_event", t.Name())
	assert.NoError(t, setMaskFile(t, resctrlDir, "3ff"))

	type args struct {
		minWaterLine, maxWaterLine, wantL3, wantMb, WantFinalL3, wantFinalMb int
		cfg                                                                  config.CacheConfig
	}
	tests := []struct {
		preHook  func(t *testing.T)
		postHook func(t *testing.T)
		name     string
		args     args
	}{
		{
			name: "TC-start dynamic",
			args: args{cfg: config.CacheConfig{
				DefaultResctrlDir: resctrlDir,
				DefaultLimitMode:  dynamicMode,
				PerfDuration:      10,
				L3Percent: config.MultiLvlPercent{
					High: 50,
					Low:  20,
					Mid:  30,
				},
				MemBandPercent: config.MultiLvlPercent{
					High: 50,
					Low:  10,
					Mid:  30,
				},
			},
				minWaterLine: 0,
				maxWaterLine: 0,
				wantL3:       20,
				wantMb:       10,
				WantFinalL3:  20,
				wantFinalMb:  10,
			},
			preHook: func(t *testing.T) {
				pi := &typedef.PodInfo{
					UID:             "abcde",
					CgroupPath:      filepath.Base(testCGRoot),
					CacheLimitLevel: lowLevel,
					Containers:      make(map[string]*typedef.ContainerInfo),
				}
				cpm.Checkpoint.Pods[pi.UID] = pi
				try.MkdirAll(testCGRoot, constant.DefaultDirMode)
				try.WriteFile(filepath.Join(testCGRoot, "tasks"), []byte(fmt.Sprint(os.Getpid())), constant.DefaultFileMode)
			},
			postHook: func(t *testing.T) {
				try.WriteFile(filepath.Join(config.CgroupRoot, "perf_event", "tasks"), []byte(fmt.Sprint(os.Getpid())), constant.DefaultFileMode)
				try.RemoveAll(testCGRoot)
				cpm.Checkpoint.Pods = make(map[string]*typedef.PodInfo)
			},
		},
		{
			name: "TC-start dynamic with very high water line",
			args: args{cfg: config.CacheConfig{
				DefaultResctrlDir: resctrlDir,
				DefaultLimitMode:  dynamicMode,
				PerfDuration:      10,
				L3Percent: config.MultiLvlPercent{
					High: 50,
					Low:  20,
					Mid:  30,
				},
				MemBandPercent: config.MultiLvlPercent{
					High: 50,
					Low:  10,
					Mid:  30,
				},
			},
				minWaterLine: math.MaxInt64,
				maxWaterLine: math.MaxInt64,
				wantL3:       25,
				wantMb:       15,
				WantFinalL3:  50,
				wantFinalMb:  50,
			},
			preHook: func(t *testing.T) {
				pi := &typedef.PodInfo{
					UID:             "abcde",
					CgroupPath:      filepath.Base(testCGRoot),
					CacheLimitLevel: lowLevel,
					Containers:      make(map[string]*typedef.ContainerInfo),
				}
				cpm.Checkpoint.Pods[pi.UID] = pi
				try.MkdirAll(testCGRoot, constant.DefaultDirMode)
				try.WriteFile(filepath.Join(testCGRoot, "tasks"), []byte(fmt.Sprint(os.Getpid())), constant.DefaultFileMode)
			},
			postHook: func(t *testing.T) {
				try.WriteFile(filepath.Join(config.CgroupRoot, "perf_event", "tasks"), []byte(fmt.Sprint(os.Getpid())), constant.DefaultFileMode)
				try.RemoveAll(testCGRoot)
				cpm.Checkpoint.Pods = make(map[string]*typedef.PodInfo)
			},
		},
		{
			name: "TC-start dynamic with low min water line",
			args: args{cfg: config.CacheConfig{
				DefaultResctrlDir: resctrlDir,
				DefaultLimitMode:  dynamicMode,
				PerfDuration:      10,
				L3Percent: config.MultiLvlPercent{
					High: 50,
					Low:  20,
					Mid:  30,
				},
				MemBandPercent: config.MultiLvlPercent{
					High: 50,
					Low:  10,
					Mid:  30,
				},
			},
				minWaterLine: 0,
				maxWaterLine: math.MaxInt64,
				wantL3:       20,
				wantMb:       10,
				WantFinalL3:  20,
				wantFinalMb:  10,
			},
			preHook: func(t *testing.T) {
				pi := &typedef.PodInfo{
					UID:             "abcde",
					CgroupPath:      filepath.Base(testCGRoot),
					CacheLimitLevel: lowLevel,
					Containers:      make(map[string]*typedef.ContainerInfo),
				}
				cpm.Checkpoint.Pods[pi.UID] = pi
				try.MkdirAll(testCGRoot, constant.DefaultDirMode)
				try.WriteFile(filepath.Join(testCGRoot, "tasks"), []byte(fmt.Sprint(os.Getpid())), constant.DefaultFileMode)
			},
			postHook: func(t *testing.T) {
				try.WriteFile(filepath.Join(config.CgroupRoot, "perf_event", "tasks"), []byte(fmt.Sprint(os.Getpid())), constant.DefaultFileMode)
				try.RemoveAll(testCGRoot)
				cpm.Checkpoint.Pods = make(map[string]*typedef.PodInfo)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preHook != nil {
				tt.preHook(t)
			}

			l3PercentDynamic = tt.args.cfg.L3Percent.Low
			mbPercentDynamic = tt.args.cfg.MemBandPercent.Low
			startDynamic(&tt.args.cfg, tt.args.maxWaterLine, tt.args.minWaterLine)
			assert.Equal(t, tt.args.wantL3, l3PercentDynamic)
			assert.Equal(t, tt.args.wantMb, mbPercentDynamic)
			for i := 0; i < 10; i++ {
				startDynamic(&tt.args.cfg, tt.args.maxWaterLine, tt.args.minWaterLine)
			}
			assert.Equal(t, tt.args.WantFinalL3, l3PercentDynamic)
			assert.Equal(t, tt.args.wantFinalMb, mbPercentDynamic)
			if tt.postHook != nil {
				tt.postHook(t)
			}
		})
	}
}

func TestClEnabled(t *testing.T) {
	oldEnbaled := enable
	tests := []struct {
		preHook  func(t *testing.T)
		postHook func(t *testing.T)
		name     string
		want     bool
	}{
		{
			name: "TC-return enabled",
			preHook: func(t *testing.T) {
				enable = true
			},
			postHook: func(t *testing.T) {
				enable = oldEnbaled
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preHook != nil {
				tt.preHook(t)
			}
			if got := ClEnabled(); got != tt.want {
				t.Errorf("ClEnabled() = %v, want %v", got, tt.want)
			}
			if tt.postHook != nil {
				tt.postHook(t)
			}
		})
	}
}

// TestDynamicExist test dynamicExist
func TestDynamicExist(t *testing.T) {
	initCpm()
	cpm.Checkpoint.Pods["podabc"].CacheLimitLevel = lowLevel
	assert.Equal(t, false, dynamicExist())
	cpm.Checkpoint.Pods["podabc"].CacheLimitLevel = dynamicLevel
	assert.Equal(t, true, dynamicExist())
}

// TestIsHostPidns test isHostPidns
func TestIsHostPidns(t *testing.T) {
	assert.Equal(t, false, isHostPidns(filepath.Join(constant.TmpTestDir, "path/not/exist/pid")))
	assert.Equal(t, true, isHostPidns("/proc/self/ns/pid"))
}

// TestInit test Init
func TestInit(t *testing.T) {
	resctrlDir := try.GenTestDir().String()
	schemataPath := filepath.Join(resctrlDir, schemataFile)
	_, err := os.Create(schemataPath)
	assert.NoError(t, err)
	assert.NoError(t, setMaskFile(t, resctrlDir, "3ff"))
	var TC1WantErr bool
	if !perf.HwSupport() {
		TC1WantErr = true
	}
	type args struct {
		cfg config.CacheConfig
	}
	tests := []struct {
		preHook  func(t *testing.T)
		postHook func(t *testing.T)
		name     string
		args     args
		wantErr  bool
	}{
		{
			name:    "TC-normal testcase",
			wantErr: TC1WantErr,
			args: args{cfg: config.CacheConfig{
				DefaultResctrlDir: resctrlDir,
				DefaultLimitMode:  dynamicMode,
				PerfDuration:      10,
				L3Percent: config.MultiLvlPercent{
					High: 100,
					Low:  10,
					Mid:  50,
				},
				MemBandPercent: config.MultiLvlPercent{
					High: 100,
					Low:  10,
					Mid:  50,
				},
				AdjustInterval: 10,
			}},
		},
		{
			name:    "TC-invalid cache config",
			wantErr: true,
			args: args{cfg: config.CacheConfig{
				AdjustInterval: 0,
			}},
		},
		{
			name:    "TC-resctrl not exist",
			wantErr: true,
			args: args{cfg: config.CacheConfig{
				DefaultResctrlDir: "/path/not/exist",
				DefaultLimitMode:  dynamicMode,
				PerfDuration:      10,
				L3Percent: config.MultiLvlPercent{
					High: 100,
					Low:  10,
					Mid:  50,
				},
				MemBandPercent: config.MultiLvlPercent{
					High: 100,
					Low:  10,
					Mid:  50,
				},
				AdjustInterval: 10,
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preHook != nil {
				tt.preHook(t)
			}
			var cfg config.CacheConfig
			cfg = tt.args.cfg
			m := &checkpoint.Manager{
				Checkpoint: &checkpoint.Checkpoint{
					Pods: make(map[string]*typedef.PodInfo),
				},
			}
			if err := Init(m, &cfg); (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.postHook != nil {
				tt.postHook(t)
			}
		})
	}
}
