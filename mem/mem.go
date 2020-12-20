package mem

import (
	"encoding/json"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/chopin11/goutil/internal/common"
	"golang.org/x/sys/unix"
)

const (
	memoryInfoProcFile = "/proc/meminfo"
)

// Memory usage statistics. Total, Available and Used contain numbers of bytes
// for human consumption.
//
// The other fields in this struct contain kernel specific values.
type VirtualMemoryStat struct {
	// Total amount of RAM on this system
	Total uint64 `json:"total"`

	// RAM available for programs to allocate
	//
	// This value is computed from the kernel specific values.
	Available uint64 `json:"available"`

	// RAM used by programs
	//
	// This value is computed from the kernel specific values.
	Used uint64 `json:"used"`

	// Percentage of RAM used by programs
	//
	// This value is computed from the kernel specific values.
	UsedPercent float64 `json:"usedPercent"`

	// This is the kernel's notion of free memory; RAM chips whose bits nobody
	// cares about the value of right now. For a human consumable number,
	// Available is what you really want.
	Free uint64 `json:"free"`

	// OS X / BSD specific numbers:
	// http://www.macyourself.com/2010/02/17/what-is-free-wired-active-and-inactive-system-memory-ram/
	Active   uint64 `json:"active"`
	Inactive uint64 `json:"inactive"`
	Wired    uint64 `json:"wired"`

	// FreeBSD specific numbers:
	// https://reviews.freebsd.org/D8467
	Laundry uint64 `json:"laundry"`

	// Linux specific numbers
	// https://www.centos.org/docs/5/html/5.1/Deployment_Guide/s2-proc-meminfo.html
	// https://www.kernel.org/doc/Documentation/filesystems/proc.txt
	// https://www.kernel.org/doc/Documentation/vm/overcommit-accounting
	Buffers        uint64 `json:"buffers"`
	Cached         uint64 `json:"cached"`
	WriteBack      uint64 `json:"writeBack"`
	Dirty          uint64 `json:"dirty"`
	WriteBackTmp   uint64 `json:"writeBackTmp"`
	Shared         uint64 `json:"shared"`
	Slab           uint64 `json:"slab"`
	Sreclaimable   uint64 `json:"sreclaimable"`
	Sunreclaim     uint64 `json:"sunreclaim"`
	PageTables     uint64 `json:"pageTables"`
	SwapCached     uint64 `json:"swapCached"`
	CommitLimit    uint64 `json:"commitLimit"`
	CommittedAS    uint64 `json:"committedAS"`
	HighTotal      uint64 `json:"highTotal"`
	HighFree       uint64 `json:"highFree"`
	LowTotal       uint64 `json:"lowTotal"`
	LowFree        uint64 `json:"lowFree"`
	SwapTotal      uint64 `json:"swapTotal"`
	SwapFree       uint64 `json:"swapFree"`
	Mapped         uint64 `json:"mapped"`
	VmallocTotal   uint64 `json:"vmallocTotal"`
	VmallocUsed    uint64 `json:"vmallocUsed"`
	VmallocChunk   uint64 `json:"vmallocChunk"`
	HugePagesTotal uint64 `json:"hugePagesTotal"`
	HugePagesFree  uint64 `json:"hugePagesFree"`
	HugePageSize   uint64 `json:"hugePageSize"`
}

type SwapMemoryStat struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"usedPercent"`
	Sin         uint64  `json:"sin"`
	Sout        uint64  `json:"sout"`
	PgIn        uint64  `json:"pgIn"`
	PgOut       uint64  `json:"pgOut"`
	PgFault     uint64  `json:"pgFault"`

	// Linux specific numbers
	// https://www.kernel.org/doc/Documentation/cgroup-v2.txt
	PgMajFault uint64 `json:"pgMajFault"`
}

type VirtualMemoryExStat struct {
	ActiveFile   uint64 `json:"activefile"`
	InactiveFile uint64 `json:"inactivefile"`
	ActiveAnon   uint64 `json:"activeanon"`
	InactiveAnon uint64 `json:"inactiveanon"`
	Unevictable  uint64 `json:"unevictable"`
}

func (m VirtualMemoryStat) String() string {
	s, _ := json.Marshal(m)
	return string(s)
}

func (m SwapMemoryStat) String() string {
	s, _ := json.Marshal(m)
	return string(s)
}

func (v VirtualMemoryExStat) String() string {
	s, _ := json.Marshal(v)
	return string(s)
}

func VirtualMemory() (*VirtualMemoryStat, error) {
	vm, _, err := readMemoryInfo()
	if err != nil {
		return nil, err
	}

	return vm, nil
}

func VirtualMemoryEx() (*VirtualMemoryExStat, error) {
	_, vmEx, err := readMemoryInfo()
	if err != nil {
		return nil, err
	}
	return vmEx, nil
}

func SwapMemory() (*SwapMemoryStat, error) {
	sysinfo := &unix.Sysinfo_t{}

	if err := unix.Sysinfo(sysinfo); err != nil {
		return nil, err
	}
	ret := &SwapMemoryStat{
		Total: uint64(sysinfo.Totalswap) * uint64(sysinfo.Unit),
		Free:  uint64(sysinfo.Freeswap) * uint64(sysinfo.Unit),
	}
	ret.Used = ret.Total - ret.Free
	//check Infinity
	if ret.Total != 0 {
		ret.UsedPercent = float64(ret.Total-ret.Free) / float64(ret.Total) * 100.0
	} else {
		ret.UsedPercent = 0
	}

	lines, _ := common.ReadLines("/proc/vmstat")
	for _, l := range lines {
		fields := strings.Fields(l)
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "pswpin":
			value, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				continue
			}
			ret.Sin = value * 4 * 1024
		case "pswpout":
			value, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				continue
			}
			ret.Sout = value * 4 * 1024
		case "pgpgIn":
			value, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				continue
			}
			ret.PgIn = value * 4 * 1024
		case "pgpgOut":
			value, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				continue
			}
			ret.PgOut = value * 4 * 1024
		case "pgFault":
			value, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				continue
			}
			ret.PgFault = value * 4 * 1024
		case "pgMajFault":
			value, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				continue
			}
			ret.PgMajFault = value * 4 * 1024
		}
	}
	return ret, nil
}

func readMemoryInfo() (*VirtualMemoryStat, *VirtualMemoryExStat, error) {
	lines, _ := common.ReadLines(memoryInfoProcFile)

	// flag if MemAvailable is in /proc/meminfo (kernel 3.14+)
	memavail := false
	activeFile := false   // "Active(file)" not available: 2.6.28 / Dec 2008
	inactiveFile := false // "Inactive(file)" not available: 2.6.28 / Dec 2008
	sReclaimable := false // "Sreclaimable:" not available: 2.6.19 / Nov 2006

	ret := &VirtualMemoryStat{}
	retEx := &VirtualMemoryExStat{}

	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) != 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])
		value = strings.Replace(value, " kB", "", -1)

		switch key {
		case "MemTotal":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.Total = t * 1024
		case "MemFree":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.Free = t * 1024
		case "MemAvailable":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			memavail = true
			ret.Available = t * 1024
		case "Buffers":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.Buffers = t * 1024
		case "Cached":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.Cached = t * 1024
		case "Active":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.Active = t * 1024
		case "Inactive":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.Inactive = t * 1024
		case "Active(anon)":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			retEx.ActiveAnon = t * 1024
		case "Inactive(anon)":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			retEx.InactiveAnon = t * 1024
		case "Active(file)":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			activeFile = true
			retEx.ActiveFile = t * 1024
		case "Inactive(file)":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			inactiveFile = true
			retEx.InactiveFile = t * 1024
		case "Unevictable":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			retEx.Unevictable = t * 1024
		case "WriteBack":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.WriteBack = t * 1024
		case "WriteBackTmp":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.WriteBackTmp = t * 1024
		case "Dirty":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.Dirty = t * 1024
		case "Shmem":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.Shared = t * 1024
		case "Slab":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.Slab = t * 1024
		case "SReclaimable":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			sReclaimable = true
			ret.Sreclaimable = t * 1024
		case "SUnreclaim":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.Sunreclaim = t * 1024
		case "PageTables":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.PageTables = t * 1024
		case "SwapCached":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.SwapCached = t * 1024
		case "CommitLimit":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.CommitLimit = t * 1024
		case "Committed_AS":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.CommittedAS = t * 1024
		case "HighTotal":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.HighTotal = t * 1024
		case "HighFree":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.HighFree = t * 1024
		case "LowTotal":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.LowTotal = t * 1024
		case "LowFree":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.LowFree = t * 1024
		case "SwapTotal":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.SwapTotal = t * 1024
		case "SwapFree":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.SwapFree = t * 1024
		case "Mapped":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.Mapped = t * 1024
		case "VmallocTotal":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.VmallocTotal = t * 1024
		case "VmallocUsed":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.VmallocUsed = t * 1024
		case "VmallocChunk":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.VmallocChunk = t * 1024
		case "HugePages_Total":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.HugePagesTotal = t
		case "HugePages_Free":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.HugePagesFree = t
		case "Hugepagesize":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, retEx, err
			}
			ret.HugePageSize = t * 1024
		}
	}

	ret.Cached += ret.Sreclaimable

	if !memavail {
		if activeFile && inactiveFile && sReclaimable {
			ret.Available = calcuateAvailVmem(ret, retEx)
		} else {
			ret.Available = ret.Cached + ret.Free
		}
	}

	ret.Used = ret.Total - ret.Free - ret.Buffers - ret.Cached
	ret.UsedPercent = float64(ret.Used) / float64(ret.Total) * 100.0

	return ret, retEx, nil
}

// calcuateAvailVmem is a fallback under kernel 3.14 where /proc/meminfo does not provide
// "MemAvailable:" column. It reimplements an algorithm from the link below
// https://github.com/giampaolo/psutil/pull/890
func calcuateAvailVmem(ret *VirtualMemoryStat, retEx *VirtualMemoryExStat) uint64 {
	var watermarkLow uint64

	lines, err := common.ReadLines("/proc/zoneinfo")
	if err != nil {
		return ret.Free + ret.Cached // fallback under kernel 2.6.13
	}

	pagesize := uint64(os.Getpagesize())
	watermarkLow = 0

	for _, line := range lines {
		fields := strings.Fields(line)

		if strings.HasPrefix(fields[0], "low") {
			lowValue, err := strconv.ParseUint(fields[1], 10, 64)

			if err != nil {
				lowValue = 0
			}
			watermarkLow += lowValue
		}
	}

	watermarkLow *= pagesize

	availMemory := ret.Free - watermarkLow
	pageCache := retEx.ActiveFile + retEx.InactiveFile
	pageCache -= uint64(math.Min(float64(pageCache/2), float64(watermarkLow)))
	availMemory += pageCache
	availMemory += ret.Sreclaimable - uint64(math.Min(float64(ret.Sreclaimable/2.0), float64(watermarkLow)))

	if availMemory < 0 {
		availMemory = 0
	}

	return availMemory
}
