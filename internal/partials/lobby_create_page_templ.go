// Code generated by templ@v0.2.364 DO NOT EDIT.

package partials

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

func LobbyCreatePage() templ.Component {
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
		_, err = templBuffer.WriteString("<!doctype html><html lang=\"en\"><head><meta charset=\"UTF-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"><script src=\"https://unpkg.com/htmx.org@1.9.6\" integrity=\"sha384-FhXw7b6AlE/jyjlZH5iHa/tTe9EpJ1Y55RjcgPbjeWMskSxZt1v9qkxLJWNJaGni\" crossorigin=\"anonymous\">")
		if err != nil {
			return err
		}
		var_2 := ``
		_, err = templBuffer.WriteString(var_2)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</script><script src=\"/bundle/js/mainLayout.js\">")
		if err != nil {
			return err
		}
		var_3 := ``
		_, err = templBuffer.WriteString(var_3)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</script><link rel=\"stylesheet\" href=\"/bundle/css/main_layout.css\"><title>")
		if err != nil {
			return err
		}
		var_4 := `Create Lobby`
		_, err = templBuffer.WriteString(var_4)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</title></head><body><nav id=\"nav_bar\" class=\"bg-blue-700 p-12  py-3  font-mono text-3xl text-white\"><a href=\"/\">")
		if err != nil {
			return err
		}
		var_5 := `Home`
		_, err = templBuffer.WriteString(var_5)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</a></nav><main id=\"main_container\" class=\"bg-blue-700 min-h-screen px-12 pt-12 pb-24\"><form class=\"bg-orange-500 p-12 flex gap-3\"><div class=\"flex flex-col mb-0.5\"><label for=\"num_of_players\" class=\"font-black\">")
		if err != nil {
			return err
		}
		var_6 := `number of players`
		_, err = templBuffer.WriteString(var_6)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</label><input type=\"number\" class=\"px-2 py-1.5 rounded-md\" name=\"num_of_players\" id=\"num_of_players\"></div><div class=\"flex flex-col mb-0.5\"><label for=\"max_hand_size\" class=\"font-black\">")
		if err != nil {
			return err
		}
		var_7 := `max hand size`
		_, err = templBuffer.WriteString(var_7)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</label><input type=\"number\" class=\"px-2 py-1.5 rounded-md\" name=\"max_hand_size\" id=\"max_hand_size\"></div></form></main></body></html>")
		if err != nil {
			return err
		}
		if !templIsBuffer {
			_, err = templBuffer.WriteTo(w)
		}
		return err
	})
}
