// Code generated by templ@v0.2.364 DO NOT EDIT.

package views

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

func LobbyPage(connectionString, lobbyId, username string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		templBuffer, templIsBuffer := w.(*bytes.Buffer)
		if !templIsBuffer {
			templBuffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templBuffer)
		}
		ctx = templ.InitializeContext(ctx)
		var_1 := templ.GetChildren(ctx)
		if var_1 == nil {
			var_1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, err = templBuffer.WriteString("<main id=\"main_container\" class=\"bg-blue-700 min-h-screen px-12 pt-12 pb-24 font-mono\" hx-ext=\"ws\" hx-swap-oob=\"beforeend\" ws-connect=\"")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString(templ.EscapeString(connectionString))
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("\"><div id=\"username\" data-username=\"")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString(templ.EscapeString(username))
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("\"></div><div id=\"lobby-id\" data-lobby-id=\"")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString(templ.EscapeString(lobbyId))
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("\"></div><div class=\"grid grid-rows-lobby_grid grid-cols-5 gap-3 h-[75vh] rounded-md\"><!--")
		if err != nil {
			return err
		}
		var_2 := ` Header row  `
		_, err = templBuffer.WriteString(var_2)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("--><div class=\"col-span-full row-span-1 bg-white flex items-center rounded-md p-5\"><h1 class=\"text-2xl\">")
		if err != nil {
			return err
		}
		var_3 := `Lobby id: `
		_, err = templBuffer.WriteString(var_3)
		if err != nil {
			return err
		}
		var var_4 string = lobbyId
		_, err = templBuffer.WriteString(templ.EscapeString(var_4))
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</h1></div><!--")
		if err != nil {
			return err
		}
		var_5 := ` Player settings `
		_, err = templBuffer.WriteString(var_5)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("--><div class=\"row-start-2 row-end-3 col-span-3 bg-white p-3 rounded-md\"><div class=\"flex gap-5\"><div class=\"h-22 bg-blue-500 rounded-full\"></div>")
		if err != nil {
			return err
		}
		err = PlayerColorComponent("green").Render(ctx, templBuffer)
		if err != nil {
			return err
		}
		err = PlayerColorComponent("blue").Render(ctx, templBuffer)
		if err != nil {
			return err
		}
		err = PlayerColorComponent("red").Render(ctx, templBuffer)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</div></div><!--")
		if err != nil {
			return err
		}
		var_6 := ` Player chat`
		_, err = templBuffer.WriteString(var_6)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("--><div class=\"row-start-2 row-end-3 col-span-2 bg-white rounded-md p-3\"><div class=\"flex flex-col h-full shadow-md\"><!--")
		if err != nil {
			return err
		}
		var_7 := ` Chat messages  `
		_, err = templBuffer.WriteString(var_7)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("--><div class=\"bg-gray-100 h-full rounded-t-md\"><div id=\"ws-events\"></div></div><!--")
		if err != nil {
			return err
		}
		var_8 := ` Chat input  `
		_, err = templBuffer.WriteString(var_8)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("--><div class=\"flex flex-grow\"><textarea ws-send hx-trigger=\"keydown[!shiftKey&amp;&amp;key==&#39;Enter&#39;]\" form=\"chat-form\" rows=\"3\" class=\"w-full max-w-full  bg-gray-200 px-1.5 py-0.5 rounded-b-md resize-none\" id=\"chat-input\" name=\"message\" type=\"text\"></textarea></div></div></div></div><script src=\"/bundle/js/lobby.js\">")
		if err != nil {
			return err
		}
		var_9 := ``
		_, err = templBuffer.WriteString(var_9)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</script></main>")
		if err != nil {
			return err
		}
		if !templIsBuffer {
			_, err = templBuffer.WriteTo(w)
		}
		return err
	})
}

func PlayerColorComponent(color string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		templBuffer, templIsBuffer := w.(*bytes.Buffer)
		if !templIsBuffer {
			templBuffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templBuffer)
		}
		ctx = templ.InitializeContext(ctx)
		var_10 := templ.GetChildren(ctx)
		if var_10 == nil {
			var_10 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		switch color {
		case "green":
			_, err = templBuffer.WriteString("<div ws-send id=\"green\" hx-trigger=\"click\" class=\"h-20 w-20 rounded-full bg-green-500 hover:shadow-md hover:shadow-green-500/50\"></div>")
			if err != nil {
				return err
			}
		case "blue":
			_, err = templBuffer.WriteString("<div ws-send id=\"blue\" hx-trigger=\"click\" class=\"h-20 w-20 rounded-full bg-blue-500 hover:shadow-md hover:shadow-blue-500/50\"></div>")
			if err != nil {
				return err
			}
		case "red":
			_, err = templBuffer.WriteString("<div ws-send id=\"red\" hx-trigger=\"click\" class=\" h-20 w-20 rounded-full bg-red-500 hover:shadow-md\n    hover:shadow-red-500/50\"></div>")
			if err != nil {
				return err
			}
		}
		if !templIsBuffer {
			_, err = templBuffer.WriteTo(w)
		}
		return err
	})
}
