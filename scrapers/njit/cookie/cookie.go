package cookie

import (
	log "github.com/Sirupsen/logrus"
	"net/http"
	"strconv"
	"sync"
)

type BakedCookie struct {
	sync.Mutex
	cookie *http.Cookie
	name   string
}

func (b *BakedCookie) Get() *http.Cookie {
	b.Lock()
	defer b.Unlock()
	return b.cookie
}

func (b *BakedCookie) SetValue(value string) {
	b.Lock()
	defer b.Unlock()
	b.cookie.Value = value
}

type cookieCutout interface {
	push(*BakedCookie)
	pop() *BakedCookie
}

func New(cookies []*http.Cookie, init func(baked *BakedCookie) error) *CookieCutter {
	cc := &CookieCutter{
		cq:   make(chan *BakedCookie, len(cookies)),
		peek: nil,
	}

	cc.wg.Add(len(cookies))

	for i := range cookies {
		bc := &BakedCookie{cookie: cookies[i], name: strconv.Itoa(i)}
		go cc.initializeCookie(bc, init)
	}

	cc.wg.Wait()

	return cc
}

func (cc *CookieCutter) initializeCookie(baked *BakedCookie, init func(baked *BakedCookie) error) {
	if err := init(baked); err != nil {
		log.Fatalln("failed to initialize cookie:", err)
	}
	cc.cq <- baked
	cc.wg.Done()
}

func (cc *CookieCutter) Push(baked *BakedCookie, onPush func(baked *BakedCookie) error) {
	if onPush != nil {
		if err := onPush(baked); err == nil {
			cc.cq <- baked
		}
	} else {
		cc.cq <- baked
	}
}

func (cc *CookieCutter) Pop(onPop func(baked *BakedCookie) error) *BakedCookie {
	bc := <-cc.cq
	if onPop != nil {
		if err := onPop(bc); err == nil {

		}
	} else {
		return bc
	}

	return nil
}

type CookieCutter struct {
	wg   sync.WaitGroup
	cq   chan *BakedCookie
	peek *BakedCookie
}
