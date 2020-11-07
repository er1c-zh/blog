---
title: "go内存分配"
date: 2020-11-06T19:11:51+08:00
draft: true
tags:
    - go
    - what
---

# 内存分配策略

根据要分配的内存大小与是否包含指针，采用了不同的分配策略。

## 极小内存分配器

要分配的内存小于等于16B*(go1.15 src/runtime/malloc.go#133)*而且不包含指针。

极小内存分配器将若干小的需求合并到一个内存块中。

## 小内存

要分配的内存小于等于32KB并不能用极小内存分配器分配的内存。

逐级尝试`mcache`，`mcentral`，`mheap`上获取`mspan`进行分配。

## 大内存

直接在`mheap`上分配。

# 在堆上分配内存

```go
// runtime/malloc.go#1194
func newobject(typ *_type/* 要分配的类型 */) unsafe.Pointer {
	return mallocgc(typ.size, typ, true)
}
```

```go
// Allocate an object of size bytes.
// Small objects are allocated from the per-P cache's free lists.
// 小对象先从每个processor的mcache上分配
// Large objects (> 32 kB) are allocated straight from the heap.
// 大对象直接从mheap上分配
func mallocgc(size uintptr, typ *_type, needzero bool) unsafe.Pointer {
	if gcphase == _GCmarktermination {
		throw("mallocgc called with gcphase == _GCmarktermination")
	}

	if size == 0 {
		return unsafe.Pointer(&zerobase)
	}

	if debug.sbrk != 0 {
		align := uintptr(16)
		if typ != nil {
			// TODO(austin): This should be just
			//   align = uintptr(typ.align)
			// but that's only 4 on 32-bit platforms,
			// even if there's a uint64 field in typ (see #599).
			// This causes 64-bit atomic accesses to panic.
			// Hence, we use stricter alignment that matches
			// the normal allocator better.
			if size&7 == 0 {
				align = 8
			} else if size&3 == 0 {
				align = 4
			} else if size&1 == 0 {
				align = 2
			} else {
				align = 1
			}
		}
		return persistentalloc(size, align, &memstats.other_sys)
	}

	// assistG is the G to charge for this allocation, or nil if
	// GC is not currently active.
	var assistG *g
	if gcBlackenEnabled != 0 {
    // 如果在gc中，需要要求分配内存的用户goroutine帮助gc
    // 避免出现gc比分配慢的情况
		// Charge the current user G for this allocation.
		assistG = getg()
		if assistG.m.curg != nil {
			assistG = assistG.m.curg
		}
		// Charge the allocation against the G. We'll account
    // for internal fragmentation at the end of mallocgc.
    // 修改该goroutine的已经帮助gc的数量
		assistG.gcAssistBytes -= int64(size)

		if assistG.gcAssistBytes < 0 {
      // 如果这个goroutine“欠”了内存，那么就要帮助gc
			// This must happen before disabling preemption.
			gcAssistAlloc(assistG)
		}
	}

	// Set mp.mallocing to keep from being preempted by GC.
	mp := acquirem()
	if mp.mallocing != 0 {
		throw("malloc deadlock")
	}
	if mp.gsignal == getg() {
		throw("malloc during signal")
	}
	mp.mallocing = 1 // 加锁，避免被GC抢占

	shouldhelpgc := false
	dataSize := size
  var c *mcache
  // 尝试获取mcache
	if mp.p != 0 {
		c = mp.p.ptr().mcache
	} else {
		// We will be called without a P while bootstrapping,
		// in which case we use mcache0, which is set in mallocinit.
		// mcache0 is cleared when bootstrapping is complete,
		// by procresize.
		c = mcache0
		if c == nil {
			throw("malloc called with no P")
		}
	}
	var span *mspan
  var x unsafe.Pointer
  // 是否要分配在不需要gc扫描的区域上（本质是是否包含指针）
	noscan := typ == nil || typ.ptrdata == 0
	if size <= maxSmallSize {
    // 极小对象的分配
		if noscan && size < maxTinySize {
      // Tiny allocator.
			//
			// Tiny allocator combines several tiny allocation requests
			// into a single memory block. The resulting memory block
			// is freed when all subobjects are unreachable. The subobjects
			// must be noscan (don't have pointers), this ensures that
			// the amount of potentially wasted memory is bounded.
			//
			// Size of the memory block used for combining (maxTinySize) is tunable.
			// Current setting is 16 bytes, which relates to 2x worst case memory
			// wastage (when all but one subobjects are unreachable).
			// 8 bytes would result in no wastage at all, but provides less
			// opportunities for combining.
			// 32 bytes provides more opportunities for combining,
			// but can lead to 4x worst case wastage.
			// The best case winning is 8x regardless of block size.
			//
			// Objects obtained from tiny allocator must not be freed explicitly.
			// So when an object will be freed explicitly, we ensure that
			// its size >= maxTinySize.
			//
			// SetFinalizer has a special case for objects potentially coming
			// from tiny allocator, it such case it allows to set finalizers
			// for an inner byte of a memory block.
			//
			// The main targets of tiny allocator are small strings and
			// standalone escaping variables. On a json benchmark
			// the allocator reduces number of allocations by ~12% and
			// reduces heap size by ~20%.
			off := c.tinyoffset
			// Align tiny pointer for required (conservative) alignment.
			if size&7 == 0 {
				off = alignUp(off, 8)
			} else if size&3 == 0 {
				off = alignUp(off, 4)
			} else if size&1 == 0 {
				off = alignUp(off, 2)
			}
			if off+size <= maxTinySize && c.tiny != 0 {
				// The object fits into existing tiny block.
				x = unsafe.Pointer(c.tiny + off)
				c.tinyoffset = off + size
				c.local_tinyallocs++
				mp.mallocing = 0
				releasem(mp)
				return x
			}
			// Allocate a new maxTinySize block.
			span = c.alloc[tinySpanClass]
			v := nextFreeFast(span)
			if v == 0 {
				v, span, shouldhelpgc = c.nextFree(tinySpanClass)
			}
			x = unsafe.Pointer(v)
			(*[2]uint64)(x)[0] = 0
			(*[2]uint64)(x)[1] = 0
			// See if we need to replace the existing tiny block with the new one
			// based on amount of remaining free space.
			if size < c.tinyoffset || c.tiny == 0 {
				c.tiny = uintptr(x)
				c.tinyoffset = size
			}
			size = maxTinySize
		} else {
      // 获取要分配的内存span大小
			var sizeclass uint8
			if size <= smallSizeMax-8 {
				sizeclass = size_to_class8[divRoundUp(size, smallSizeDiv)]
			} else {
				sizeclass = size_to_class128[divRoundUp(size-smallSizeMax, largeSizeDiv)]
			}
			size = uintptr(class_to_size[sizeclass])
      spc := makeSpanClass(sizeclass, noscan)
      // 从per-P的mcache中查找对应大小的mspan实例
      span = c.alloc[spc]
      // 尝试快速获取
			v := nextFreeFast(span)
			if v == 0 {
        // 慢速路径，会尝试逐级向上申请mcache->mcentral->mheap
				v, span, shouldhelpgc = c.nextFree(spc)
			}
			x = unsafe.Pointer(v)
			if needzero && span.needzero != 0 {
				memclrNoHeapPointers(unsafe.Pointer(v), size)
			}
		}
	} else {
		shouldhelpgc = true
		systemstack(func() {
      // 大对象的申请，直接走mheap申请
			span = largeAlloc(size, needzero, noscan)
		})
		span.freeindex = 1
		span.allocCount = 1
		x = unsafe.Pointer(span.base())
		size = span.elemsize
	}

	var scanSize uintptr
	if !noscan {
		// If allocating a defer+arg block, now that we've picked a malloc size
		// large enough to hold everything, cut the "asked for" size down to
		// just the defer header, so that the GC bitmap will record the arg block
		// as containing nothing at all (as if it were unused space at the end of
		// a malloc block caused by size rounding).
		// The defer arg areas are scanned as part of scanstack.
		if typ == deferType {
			dataSize = unsafe.Sizeof(_defer{})
		}
		heapBitsSetType(uintptr(x), size, dataSize, typ)
		if dataSize > typ.size {
			// Array allocation. If there are any
			// pointers, GC has to scan to the last
			// element.
			if typ.ptrdata != 0 {
				scanSize = dataSize - typ.size + typ.ptrdata
			}
		} else {
			scanSize = typ.ptrdata
		}
		c.local_scan += scanSize
	}

	// Ensure that the stores above that initialize x to
	// type-safe memory and set the heap bits occur before
	// the caller can make x observable to the garbage
	// collector. Otherwise, on weakly ordered machines,
	// the garbage collector could follow a pointer to x,
  // but see uninitialized memory or stale heap bits.
  // 增加屏障，避免在初始化好新的对象之前，将对象暴露给垃圾回收期。
	publicationBarrier()

  // Allocate black during GC.
  // 在GC期间，标记为黑色
  // All slots hold nil so no scanning is needed.
  // 因为所有的槽都是空的
	// This may be racing with GC so do it atomically if there can be
	// a race marking the bit.
	if gcphase != _GCoff {
		gcmarknewobject(span, uintptr(x), size, scanSize)
	}

	if raceenabled {
		racemalloc(x, size)
	}

	if msanenabled {
		msanmalloc(x, size)
	}

	mp.mallocing = 0
  // 释放锁
	releasem(mp)

	if debug.allocfreetrace != 0 {
		tracealloc(x, size, typ)
	}

	if rate := MemProfileRate; rate > 0 {
		if rate != 1 && size < c.next_sample {
			c.next_sample -= size
		} else {
			mp := acquirem()
      profilealloc(mp, x, size)
			releasem(mp)
		}
	}

	if assistG != nil {
		// Account for internal fragmentation in the assist
		// debt now that we know it.
		assistG.gcAssistBytes -= int64(size - dataSize)
	}

	if shouldhelpgc {
		if t := (gcTrigger{kind: gcTriggerHeap}); t.test() {
			gcStart(t)
		}
	}

	return x
}
```

# todo 数据结构
